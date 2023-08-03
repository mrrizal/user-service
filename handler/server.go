package handler

import "github.com/SawitProRecruitment/UserService/repository"

type Server struct {
	Repository repository.RepositoryInterface
	Service    Service
}

type NewServerOptions struct {
	Repository repository.RepositoryInterface
	Service    Service
}

func NewServer(opts NewServerOptions) *Server {
	optsValidator := NewValidatorOptions{
		Repository: opts.Repository,
	}

	optsService := NewServiceOptions{
		Repository: opts.Repository,
		Validator:  *NewValidator(optsValidator),
	}

	service := NewService(optsService)
	return &Server{opts.Repository, service}
}
