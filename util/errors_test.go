package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/angelapytao/diffgram-go/util"
)

func TestAppError_Error(t *testing.T) {
	err := util.ErrNotFound
	assert.Equal(t, "[1001] not found", err.Error())
}

func TestAppError_IsComparison(t *testing.T) {
	assert.Equal(t, 404, util.ErrNotFound.HTTPStatus)
	assert.Equal(t, 401, util.ErrUnauthorized.HTTPStatus)
	assert.Equal(t, 401, util.ErrWrongPassword.HTTPStatus)
}
