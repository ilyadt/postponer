package core

import "postponer/model"

type Dispatcher interface {
	Dispatch(message model.Message) error
}
