package rbac

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRbac(t *testing.T) {
	t.Run(`pathToRegex check`, func(t *testing.T) {
		path, method, err := parseSwaggerPattern("/api/v1/admin_panel/{id}/login [post]")
		require.Nil(t, err)
		require.Equal(t, POST, method)
		r1 := pathToRegex(path)

		validUri := "/api/v1/admin_panel/123-321/login"
		isMatch := r1.MatchString(validUri)
		require.Equal(t, true, isMatch)

		invalidUri := "/api/v1/admin_panel/login"
		isMatch = r1.MatchString(invalidUri)
		require.Equal(t, false, isMatch)

		path, method, err = parseSwaggerPattern("/api/v1/admin_panel/{id}/login/{otherID} [post]")
		require.Nil(t, err)
		require.Equal(t, POST, method)
		r2 := pathToRegex(path)

		validUri = "/api/v1/admin_panel/123-321/login/qwe-ewr123-wr-12"
		isMatch = r2.MatchString(validUri)
		require.Equal(t, true, isMatch)

		invalidUri = "/api/v1/admin_panel/we-ewr123-wr-12/login"
		isMatch = r2.MatchString(invalidUri)
		require.Equal(t, false, isMatch)
	})

}
