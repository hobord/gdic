package gdic

import (
	"sync"
)

// InstanceName is a name of instance for store in the container
type InstanceName string

const (
	Default InstanceName = "default"
	// you can define more instance name in Your project
)

// store is a module level singleton container
var store container

// container is store instances and factories
type container struct {
	// Lck is a mutex for locking the container
	Lck sync.RWMutex

	// instances is storage for the created instances
	instances map[string]map[InstanceName]interface{}
	// factories is storage for the factories for creating instances into the container
	factories map[string]map[InstanceName]difactory
}

// create module level singleton container
func init() {
	store = container{
		instances: make(map[string]map[InstanceName]interface{}),
		factories: make(map[string]map[InstanceName]difactory),
	}
}
