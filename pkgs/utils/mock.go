package utils

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type responseWriterMock struct{}

func (w *responseWriterMock) Header() http.Header {
	return http.Header{}
}

func (w *responseWriterMock) Write([]byte) (int, error) {
	return 0, nil
}

func (w *responseWriterMock) WriteHeader(statusCode int) {}

func GetWorkerGinContext() *gin.Context {
	req, _ := http.NewRequest("POST", "/", nil)
	req = req.WithContext(context.Background())

	w := &responseWriterMock{}
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	return c
}
