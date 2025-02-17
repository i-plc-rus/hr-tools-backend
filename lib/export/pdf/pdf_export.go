package pdfexport

import (
	"bytes"
	"fmt"
	"github.com/go-pdf/fpdf"
	"github.com/pkg/errors"
	"hr-tools-backend/models"
	"html/template"
	"strings"
)

func GenerateOffer(pdfOfferTemplate string, tplData models.TemplateData) (pdfFile []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("GenerateOffer panic recover: %v", r)
		}
	}()
	pdf := fpdf.New("P", "mm", "A4", "static/font/")
	pdf.AddPage()
	pdf.AddUTF8Font("Arial", "", "Arial.ttf")
	pdf.AddUTF8Font("Arial", "B", "Arial Bold.ttf")
	pdf.AddUTF8Font("Arial", "I", "Arial Italic.ttf")
	pdf.AddUTF8Font("Arial", "BI", "Arial Bold Italic.ttf")
	pdf.SetFont("Arial", "", 14)
	if pdf.Error() != nil {
		return nil, pdf.Error()
	}
	err = putImg(pdf, tplData.Files.Logo)
	if err != nil {
		return nil, err
	}
	err = putImg(pdf, tplData.Files.Sign)
	if err != nil {
		return nil, err
	}
	err = putImg(pdf, tplData.Files.Stamp)
	if err != nil {
		return nil, err
	}

	// лого заголовок
	if tplData.Files.Logo != nil {
		pdf.Image(tplData.Files.Logo.FileName, 10, 12, 30, 0, false, "", 0, "")
	}

	pdf.SetLeftMargin(45)
	_, lineHt := pdf.GetFontSize()
	htmlStr := fmt.Sprintf("%v<br>", tplData.CompanyName) +
		fmt.Sprintf("%v<br>", tplData.CompanyContact) +
		fmt.Sprintf("%v<br>", tplData.CompanyAddress)
	html := pdf.HTMLBasicNew()
	html.Write(lineHt, htmlStr)
	pdf.SetLeftMargin(10)

	posY := pdf.GetY()
	if posY < 50 {
		posY = 50
		pdf.SetY(posY)
	}

	// текст
	_, lineHt = pdf.GetFontSize()

	tpl, err := template.New("offer_body").Parse(pdfOfferTemplate)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, tplData)
	if err != nil {
		return nil, err
	}
	htmlStr = buf.String()

	html = pdf.HTMLBasicNew()
	html.Write(lineHt, htmlStr)
	posY = pdf.GetY()
	posY += 10
	if tplData.Files.Stamp != nil {
		pdf.Image(tplData.Files.Stamp.FileName, 30, posY, 30, 0, false, "", 0, "")
	}
	pageX, _, _ := pdf.PageSize(1)
	if tplData.Files.Sign != nil {
		pdf.Image(tplData.Files.Sign.FileName, pageX-50, posY, 30, 0, false, "", 0, "")
	}

	buf = new(bytes.Buffer)
	err = pdf.Output(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
	// fileStr := "offer.pdf"
	// return pdf.OutputFileAndClose(fileStr)
}

func putImg(pdf *fpdf.Fpdf, fileData *models.File) (err error) {
	if fileData == nil {
		return nil
	}
	options := fpdf.ImageOptions{
		ReadDpi:   false,
		ImageType: "",
	}

	if options.ImageType == "" {
		options.ImageType, err = GetImgType(fileData.FileName)
		if err != nil {
			return err
		}
	}
	reader := bytes.NewReader(fileData.Body)
	pdf.RegisterImageOptionsReader(fileData.FileName, options, reader)
	return pdf.Error()
}

func GetImgType(fileName string) (string, error) {
	pos := strings.LastIndex(fileName, ".")
	if pos < 0 {
		return "", errors.Errorf("не удалось получить расширение файла: %s", fileName)
	}
	return fileName[pos+1:], nil
}
