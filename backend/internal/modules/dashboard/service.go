package dashboard

type Counter func() int

type Service struct {
	users        Counter
	activities   Counter
	applications Counter
}

type Overview struct {
	UserCount        int `json:"user_count"`
	ActivityCount    int `json:"activity_count"`
	ApplicationCount int `json:"application_count"`
}

func NewService(users Counter, activities Counter, applications Counter) *Service {
	return &Service{
		users:        users,
		activities:   activities,
		applications: applications,
	}
}

func (s *Service) Overview() Overview {
	overview := Overview{}
	if s.users != nil {
		overview.UserCount = s.users()
	}
	if s.activities != nil {
		overview.ActivityCount = s.activities()
	}
	if s.applications != nil {
		overview.ApplicationCount = s.applications()
	}

	return overview
}
