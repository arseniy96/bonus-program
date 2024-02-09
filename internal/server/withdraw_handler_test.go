package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/arseniy96/bonus-program/internal/mocks"
	"github.com/arseniy96/bonus-program/internal/store"
)

func TestServer_WithdrawHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepository(ctrl)
	m.EXPECT().FindUserByToken(gomock.Any(), allowedTokenHash).Return(&store.User{ID: 1, Bonuses: 100000}, nil)
	m.EXPECT().FindUserByToken(gomock.Any(), allowedToken2Hash).Return(&store.User{ID: 2, Bonuses: 10000}, nil)
	m.EXPECT().FindUserByToken(gomock.Any(), wrongTokenHash).Return(nil, fmt.Errorf("invalid token"))
	m.EXPECT().SaveWithdrawBonuses(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	type fields struct {
		Repository Repository
		AuthToken  string
		body       string
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
				body:       `{"order":"123","sum":500}`,
			},
			want: results{
				statusCode: http.StatusOK,
				response:   `{}`,
			},
		},
		{
			name: "invalid auth token",
			fields: fields{
				Repository: m,
				AuthToken:  wrongToken,
				body:       `{"order":"123","sum":500}`,
			},
			want: results{
				statusCode: http.StatusInternalServerError,
				response:   ``,
			},
		},
		{
			name: "insufficient funds",
			fields: fields{
				Repository: m,
				AuthToken:  allowedToken2,
				body:       `{"order":"123","sum":500}`,
			},
			want: results{
				statusCode: http.StatusPaymentRequired,
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
			r.POST("/api/user/balance/withdraw", s.WithdrawHandler)
			req, _ := http.NewRequest("POST", "/api/user/balance/withdraw", strings.NewReader(tt.fields.body))
			w := httptest.NewRecorder()

			req.Header.Set("Authorization", tt.fields.AuthToken)
			r.ServeHTTP(w, req)

			responseData, _ := io.ReadAll(w.Body)
			assert.Equal(t, tt.want.response, string(responseData))
			assert.Equal(t, tt.want.statusCode, w.Code)
		})
	}
}
