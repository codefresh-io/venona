package operators

const (
	RuntimeEnvironmentOperatorType = "runtime-environment"
	VenonaOperatorType             = "venona"
	VolumeProvisionerOperatorType  = "volume-provisioner"
	DefaultStorageClassNamePrefix  = "dind-local-volumes-venona"
)

type (
	Operator interface {
		Install() error
		Status() ([][]string, error)
		Delete() error
		Upgrade() error
	}
)

func GetOperator(t string) Operator {
	if t == VenonaOperatorType {
		return &venonaOperator{}
	}

	if t == RuntimeEnvironmentOperatorType {
		return &RuntimeEnvironmentOperator{}
	}

	if t == VolumeProvisionerOperatorType {
		return &VolumeProvisionerOperator{}
	}

	return nil
}
