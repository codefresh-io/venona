package operators

const (
	RuntimeEnvironmentOperatorType = "runtime-environment"
	VenonaOperatorType             = "venona"
)

type (
	Operator interface {
		Install() error
		Status() ([][]string, error)
		Delete() error
	}
)

func GetOperator(t string) Operator {
	if t == VenonaOperatorType {
		return &venonaOperator{}
	}

	if t == RuntimeEnvironmentOperatorType {
		return &RuntimeEnvironmentOperator{}
	}

	return nil
}
