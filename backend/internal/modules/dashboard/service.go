package dashboard

type Counter func() int
type CounterRepository interface {
	Count() int
}

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

func NewService(users any, activities any, applications any) *Service {
	return &Service{
		users:        normalizeCounter(users),
		activities:   normalizeCounter(activities),
		applications: normalizeCounter(applications),
	}
}

func normalizeCounter(source any) Counter {
	switch v := source.(type) {
	case nil:
		return nil
	case Counter:
		return v
	case func() int:
		return Counter(v)
	case CounterRepository:
		return v.Count
	default:
		return nil
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
