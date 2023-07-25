package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/arseniy96/bonus-program/internal/mocks"
	"github.com/arseniy96/bonus-program/internal/store"
)

func TestServer_GetUserBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepository(ctrl)
	m.EXPECT().FindUserByToken(gomock.Any(), allowedTokenHash).Return(&store.User{ID: 1, Bonuses: 100}, nil)
	m.EXPECT().FindUserByToken(gomock.Any(), wrongTokenHash).Return(nil, fmt.Errorf("invalid token"))
	m.EXPECT().GetWithdrawalSumByUserID(gomock.Any(), 1).Return(500, nil)

	type fields struct {
		Repository Repository
		AuthToken  string
	}
	type results struct {
		statusCode int
		response   string
	}
	tests := []struct {
		name   string
		fields fields
		want   results
	}{
		{
			name: "success response",
			fields: fields{
				Repository: m,
				AuthToken:  allowedToken,
			},
			want: results{
				statusCode: http.StatusOK,
				response:   `{"current":1,"withdrawn":5}`,
			},
		},
		{
			name: "invalid token",
			fields: fields{
				Repository: m,
				AuthToken:  wrongToken,
			},
			want: results{
				statusCode: http.StatusInternalServerError,
				response:   ``,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Repository: tt.fields.Repository,
			}

			r := SetUpRouter()
			r.GET("/api/user/balance", s.GetUserBalance)
			req, _ := http.NewRequest("GET", "/api/user/balance", nil)
			w := httptest.NewRecorder()

			req.Header.Set("Authorization", tt.fields.AuthToken)
			r.ServeHTTP(w, req)

			responseData, _ := io.ReadAll(w.Body)
			assert.Equal(t, tt.want.response, string(responseData))
			assert.Equal(t, tt.want.statusCode, w.Code)
		})
	}
}
