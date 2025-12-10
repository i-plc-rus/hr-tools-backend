package externalservices

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtHelpers(t *testing.T) {
	t.Run(`ExtractAuditData from pure ctx check`, func(t *testing.T) {
		ctx := context.TODO()
		ctxData := ExtractAuditData(ctx)
		require.Equal(t, false, ctxData.WithAudit)
	})

	t.Run(`ExtractAuditData from filled ctx check`, func(t *testing.T) {
		expectedSpaceID := "someSpaceID"
		expectedRecID := "someRecID"
		expectedUri := "someUri"
		expectedRequest := "someRequest"
		ctx := context.TODO()
		ctx = GetContextWithRecID(ctx, expectedSpaceID, expectedRecID)
		ctxData := ExtractAuditData(ctx)
		require.Equal(t, false, ctxData.WithAudit)
		require.Equal(t, expectedSpaceID, ctxData.SpaceID)
		require.Equal(t, expectedRecID, ctxData.RecID)

		ctx = GetAuditContext(ctx, expectedUri, []byte(expectedRequest))
		ctxData = ExtractAuditData(ctx)
		require.Equal(t, true, ctxData.WithAudit)
		require.Equal(t, expectedSpaceID, ctxData.SpaceID)
		require.Equal(t, expectedUri, ctxData.Uri)
		require.Equal(t, expectedRequest, ctxData.Request)
		require.Equal(t, expectedRecID, ctxData.RecID)
	})

}
