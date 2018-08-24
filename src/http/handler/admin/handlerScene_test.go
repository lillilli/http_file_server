package admin_test

import (
	"github.com/stretchr/testify/mock"
	test "github.com/lillilli/dr_web_exercise/src/http/handler/test"
	"github.com/lillilli/dr_web_exercise/src/mocks"
)

type AdminHandlerScene struct {
	test.HandlerScene
	UserTokenManagerMock *mocks.UserTokensManager
	TokenCountersManager *mocks.TokenCountersManager
}

func NewAdminHandlerScene() *AdminHandlerScene {
	s := &AdminHandlerScene{}
	s.Init()
	return s
}

func (s *AdminHandlerScene) Init() {
	s.HandlerScene.Init()

	s.UserTokenManagerMock = &mocks.UserTokensManager{}
	s.DB.On("UserTokens").Return(s.UserTokenManagerMock)

	s.TokenCountersManager = &mocks.TokenCountersManager{}
	s.DB.On("TokenCounters").Return(s.TokenCountersManager)

	s.Stats.On("sendCountByRequestType", mock.Anything, mock.Anything, mock.Anything).Return()
}
