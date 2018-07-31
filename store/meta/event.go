package meta

type Operation int8

const (
	OperationNone = Operation(iota)
	OperationCreate
	OperationUpdate
	OperationDelete
)

func (op Operation) String() string {
	return Operation_name[op]
}

var Operation_name = map[Operation]string{
	OperationNone:   "None",
	OperationCreate: "Create",
	OperationUpdate: "Update",
	OperationDelete: "Delete",
}

type Serializable interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	GetId() string
	Valid() error
}

type EventListener interface {
	RecvHost(op Operation, data *Host)
	RecvAuth(op Operation, data *Auth)
	RecvRoute(op Operation, data *Route)
	RecvService(op Operation, data *Service)
	RecvServiceConfig(op Operation, service string, data *ServiceConfig)
	RecvApi(op Operation, service string, data *Api)
	RecvServer(op Operation, service string, data *Server)
}
