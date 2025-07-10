package supersethandler

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hr-tools-backend/config"
	"hr-tools-backend/models"
	supersetapimodels "hr-tools-backend/models/api/superset"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Provider interface {
	GetGuestToken(ctx context.Context, spaceID, dashboardCode string) (token, dashboardID, hMsg string, err error)
}

var Instance Provider

type impl struct {
	host              string
	username          string
	password          string
	cache             *cache.Cache
	dashboardParamMap map[string]string //[code]dashboardID
}

func NewHandler(host, username, password string, dashboardParams models.DashboardParams) {
	dashboardParamMap := map[string]string{}
	for _, param := range dashboardParams {
		dashboardParamMap[param.Code] = param.DashboardID
	}
	Instance = impl{
		host:              host,
		username:          username,
		password:          password,
		cache:             cache.New(cacheTTL, cacheTTL),
		dashboardParamMap: dashboardParamMap,
	}
}

const (
	loginPath           string = "%v/api/v1/security/login"
	guestTokenPath      string = "%v/api/v1/security/guest_token/"
	importDashboardPath string = "%v/api/v1/dashboard/import/"
	csrfTokenPath       string = "%v/api/v1/security/csrf_token/"
	dashboardCheckPath  string = "%v/api/v1/dashboard/%v"

	provider      string = "db"
	guestUsername string = "guest_user"

	cacheKeyPattern string = "superset-guest-token:%v"
	cacheTTL               = time.Hour
	reportCacheTTL         = 5 * time.Minute
)

func getCacheKey(spaceID string) string {
	return fmt.Sprintf(cacheKeyPattern, spaceID)
}

func (i impl) GetGuestToken(ctx context.Context, spaceID, dashboardCode string) (token, dashboardID, hMsg string, err error) {
	dashboardID, ok := i.dashboardParamMap[dashboardCode]
	if !ok || dashboardID == "" {
		return "", "", "дашборд не найден", nil
	}
	cacheKey := getCacheKey(spaceID)
	cacheValue, ok := i.cache.Get(cacheKey)
	if ok {
		return cacheValue.(string), dashboardID, "", nil
	}

	accessToken, csrfToken, csrfCookies, err := i.authorize()
	if err != nil {
		return "", "", "", err
	}

	guestToken, err := i.guestToken(accessToken, csrfToken, csrfCookies, spaceID)
	if err != nil {
		return "", "", "", errors.Wrap(err, "ошибка получения гостевого токена")
	}
	i.cache.Set(cacheKey, guestToken, reportCacheTTL)
	return guestToken, dashboardID, "", nil
}

func (i impl) authorize() (accessToken, csrfToken string, csrfCookies []*http.Cookie, err error) {
	if i.password == "" {
		return "", "", nil, errors.New("не задан пароль доступа в Superset")
	}

	accessToken, err = i.accessToken()
	if err != nil {
		return "", "", nil, errors.Wrap(err, "ошибка получения токена авторизации")
	}

	csrfToken, csrfCookies, err = i.csrfToken(accessToken)
	if err != nil {
		return "", "", nil, errors.Wrap(err, "ошибка получения CSRF токена")
	}
	return accessToken, csrfToken, csrfCookies, err
}

func (i impl) accessToken() (string, error) {
	url := fmt.Sprintf(loginPath, i.host)
	loginReq := supersetapimodels.SupersetLoginReq{
		Username: i.username,
		Password: i.password,
		Provider: provider,
		Refresh:  true,
	}

	body, err := json.Marshal(loginReq)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации запроса login для superset")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", errors.Wrap(err, "ошибка запроса получения токена доступа")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "ошибка выполнения запроса логина в суперсет")
	}
	if resp.StatusCode != http.StatusOK {
		var responseBody map[string]interface{}
		if err = json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			log.Warn("ошибка чтения тела ответа в запросе логина")
		}
		return "", errors.Wrap(err, fmt.Sprintf("не удалось получить токен доступа, статус ответа %d, ошибка %v", resp.StatusCode, responseBody))
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "ошибка чтения массива байтов")
	}
	token := struct {
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
	}{}
	err = json.Unmarshal([]byte(body), &token)
	if err != nil {
		return "", errors.Wrap(err, "ошибка декодирования json в структуру в запросе логина в суперсет")
	}
	return token.Access, nil
}

func (i impl) guestToken(accessToken, csrfToken string, csrfCookies []*http.Cookie, spaceID string) (string, error) {
	url := fmt.Sprintf(guestTokenPath, i.host)
	guestReq := supersetapimodels.SupersetGuestTokenReq{
		Resources: []supersetapimodels.Resource{},
		RLS: []supersetapimodels.RLS{
			{Clause: fmt.Sprintf("space_id = '%s'", spaceID)},
		},
		User: supersetapimodels.User{
			Username: guestUsername,
		},
	}
	for _, id := range i.dashboardParamMap {
		guestReq.Resources = append(guestReq.Resources,
			supersetapimodels.Resource{
				ID:   id,
				Type: config.Conf.Superset.ResourcesType,
			})
	}

	body, err := json.Marshal(guestReq)
	if err != nil {
		return "", errors.Wrap(err, "ошибка десериализации запроса на получение гостевого токена")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", errors.Wrap(err, "ошибка запроса получения гостевого токена")
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-CSRFToken", csrfToken)
	if csrfCookies != nil {
		req.AddCookie(csrfCookies[0])
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "ошибка выполнения запроса получения гостевого токена")
	}
	if resp.StatusCode != http.StatusOK {
		var responseBody map[string]interface{}
		if err = json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			log.WithError(err).Warn("ошибка чтения тела ответа запроса гостевого токена")
		}
		return "", errors.New(fmt.Sprintf("не удалось получить гостевой токен, статус ответа %d, ошибка %v", resp.StatusCode, responseBody))
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "ошибка чтения тела ответа запроса гостевого токена")
	}
	guestToken := struct {
		Token string `json:"token"`
	}{}
	err = json.Unmarshal([]byte(body), &guestToken)
	if err != nil {
		return "", errors.Wrap(err, "ошибка декодирования json в структуру в запросе гостевого токена")
	}
	return guestToken.Token, nil
}

func (i impl) csrfToken(accessToken string) (string, []*http.Cookie, error) {
	url := fmt.Sprintf(csrfTokenPath, i.host)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", nil, errors.Wrap(err, "ошибка запроса получения csrf токена")
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, errors.Wrap(err, "ошибка выполнения запроса получения csrf токена")
	}
	if resp.StatusCode != http.StatusOK {
		var responseBody map[string]interface{}
		if err = json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			log.Warn("ошибка чтения тела ответа")
		}
		return "", nil, errors.Wrap(err, fmt.Sprintf("не удалось получить csrf токен, статус ответа %d, ошибка %v", resp.StatusCode, responseBody))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, errors.Wrap(err, "ошибка чтения тела ответа в запросе csrf токена")
	}
	csrfToken := struct {
		CsrfToken string `json:"result"`
	}{}
	err = json.Unmarshal([]byte(body), &csrfToken)
	if err != nil {
		return "", nil, errors.Wrap(err, "ошибка декодирования json в структуру в запросе csrf токена")
	}
	return csrfToken.CsrfToken, resp.Cookies(), nil
}

func (i impl) checkDashboard(accessToken, spaceID string) (bool, error) {
	url := fmt.Sprintf(dashboardCheckPath, i.host, spaceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		if responseBody != nil {
			return false, errors.Errorf("статус ответа %d, ответ: %v", resp.StatusCode, string(responseBody))
		}
		return false, errors.Errorf("статус ответа %d", resp.StatusCode)
	}
	return true, nil
}

func (i impl) importDashboard(ctx context.Context, accessToken, csrfToken string, csrfCookies []*http.Cookie, spaceID string) error {
	zipData, err := zipWriter(spaceID)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(importDashboardPath, i.host)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("formData", "output.zip")
	if err != nil {
		return err
	}

	_, err = part.Write(zipData)
	if err != nil {
		return err
	}
	err = writer.WriteField("overwrite", "true")
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return errors.Wrap(err, "ошибка создания запроса")
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-CSRFToken", csrfToken)
	if csrfCookies != nil {
		req.AddCookie(csrfCookies[0])
	}
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "ошибка выполнения запроса логина в суперсет")
	}
	responseBody, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(response.Body)
		if len(responseBody) != 0 {
			return errors.New(string(responseBody))
		}
		return errors.New(fmt.Sprintf("Status: %v", response.Status))
	}
	log.
		WithField("space_id", spaceID).
		WithField("response", string(responseBody)).
		Info("superset - создан дашборд на основе шаблона")
	return nil
}

func zipWriter(spaceID string) ([]byte, error) {
	buf := new(bytes.Buffer)

	w := zip.NewWriter(buf)
	zipping := func() error {
		err := addMeta(w, "static/dashboard_template/metadata.yaml")
		if err != nil {
			return err
		}
		err = addFolder(w, "static/dashboard_template/charts", "dashboard_template/charts", spaceID)
		if err != nil {
			return err
		}
		err = addFolder(w, "static/dashboard_template/dashboards", "dashboard_template/dashboards", spaceID)
		if err != nil {
			return err
		}
		err = addFolder(w, "static/dashboard_template/databases", "dashboard_template/databases", spaceID)
		if err != nil {
			return err
		}
		err = addFolder(w, "static/dashboard_template/datasets/PostgreSQL", "dashboard_template/datasets/PostgreSQL", spaceID)
		if err != nil {
			return err
		}
		return nil
	}

	err := zipping()
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	// err = os.WriteFile("static/output.zip", buf.Bytes(), os.ModePerm)
	// if err != nil {
	// 	return nil, err
	// }
	return buf.Bytes(), nil
}

func addMeta(w *zip.Writer, tplPath string) error {
	body, err := os.ReadFile(tplPath)
	if err != nil {
		return err
	}

	f, err := w.Create("dashboard_template/metadata.yaml")
	if err != nil {
		return err
	}
	file := bytes.NewReader(body)
	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}
	return nil
}

func addFolder(w *zip.Writer, tplDir, zipDir string, spaceID string) error {
	d, err := os.Open(tplDir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return errors.Wrapf(err, "ошибка чтения каталога c шаблонами %v", tplDir)
	}
	for _, name := range names {
		filePath := filepath.Join(tplDir, name)
		//интересуют только файлы *.yaml
		if isDir(filePath) || !strings.HasSuffix(filePath, ".yaml") {
			continue
		}
		body, err := os.ReadFile(filePath)
		if err != nil {
			return errors.Wrapf(err, "ошибка чтения файла шаблона %v", filePath)
		}
		//Замены в шаблоне
		fData := string(body)
		fData = strings.ReplaceAll(fData, "{{SPACE_UUID}}", spaceID)
		body = []byte(fData)

		zipPath := filepath.Join(zipDir, name)
		f, err := w.Create(zipPath)
		if err != nil {
			return errors.Wrapf(err, "ошибка создания файла шаблона в архиве: %v", zipPath)
		}
		file := bytes.NewReader(body)
		_, err = io.Copy(f, file)
		if err != nil {
			return errors.Wrapf(err, "ошибка сохранения файла шаблона в архив: %v", zipPath)
		}
	}
	return nil
}

func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}
