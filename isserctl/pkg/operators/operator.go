package operators

const (
	RuntimeEnvironmentOperatorType = "runtime-environment"
	IsserOperatorType              = "isser"
)

type (
	Operator interface {
		Install() error
		Status() (TableRows, error)
		Delete() error
	}
)

func GetOperator(t string) Operator {
	if t == IsserOperatorType {
		return &IsserOperator{}
	}

	if t == RuntimeEnvironmentOperatorType {
		return &RuntimeEnvironmentOperator{}
	}

	return nil
}
