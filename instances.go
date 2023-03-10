package gdic

import "reflect"

type ResolveOption func(*resolveOptions)

type resolveOptions struct {
	// IsSingleton is a flag for singleton instance
	IsSingleton    bool // default: true
	FactoryOptions []interface{}
}

// ResolveWithNoSingleton is a resolve option for resolve instance without singleton
func ResolveWithNoSingleton() ResolveOption {
	return func(o *resolveOptions) {
		o.IsSingleton = false
	}
}

// WithFactoryOptions is a resolve option for resolve instance with factory options
func WithFactoryOptions(opts ...interface{}) ResolveOption {
	return func(o *resolveOptions) {
		o.FactoryOptions = opts
	}
}

// Resolve returns instance of the interface
func Resolve[T any](name InstanceName, options ...ResolveOption) (T, error) {
	// set default options
	opts := resolveOptions{
		IsSingleton: true,
	}

	// apply options
	for _, opt := range options {
		opt(&opts)
	}

	// if options IsSingleton is true, try to get instance from the container
	if opts.IsSingleton {
		return resolveSingleton[T](name, opts)
	}

	// if options IsSingleton false, resolve instance
	return resolve[T](name, opts)
}

func resolveSingleton[T any](name InstanceName, options resolveOptions) (T, error) {
	var (
		err error
	)

	// get type of the interface
	itype := GetType[T]()

	// check if instance type is exist in the container
	store.Lck.Lock()
	if _, ok := store.instances[itype]; !ok {
		store.instances[itype] = make(map[InstanceName]interface{})
	}
	store.Lck.Unlock()

	// try to get instance from the container
	store.Lck.RLock()
	resolvedInstance, ok := store.instances[itype][name]
	store.Lck.RUnlock()

	if !ok {
		resolvedInstance, err = resolve[T](name, options)
		if err != nil {
			return resolvedInstance.(T), err
		}

		// store instance in the container
		store.Lck.Lock()
		store.instances[itype][name] = resolvedInstance
		store.Lck.Unlock()
	}

	return resolvedInstance.(T), nil
}

func resolve[T any](name InstanceName, options resolveOptions) (T, error) {
	var (
		instance T
		err      error
	)

	// get type of the interface
	itype := GetType[T]()

	// get factory from the container
	store.Lck.RLock()
	factory, ok := store.factories[itype][name]
	store.Lck.RUnlock()

	if !ok {
		return instance, ErrFactoryNotFound
	}

	// create instance
	resolvedInstance, err := factory(options.FactoryOptions...)
	if err != nil {
		return instance, err
	}

	return resolvedInstance.(T), err
}

// IsInstanceExist checks if instance is exist in the container
func IsInstanceExist[T any](name InstanceName) bool {
	// get type of the interface
	itype := GetType[T]()

	// lock the container
	store.Lck.RLock()
	defer store.Lck.RUnlock()
	// check if type is exist in the container

	if _, ok := store.instances[itype]; !ok {
		return false
	}

	// check if instance is exist
	_, ok := store.instances[itype][name]

	return ok
}

// AddInstance adds instance to the container
func AddInstance[T any](name InstanceName, instance T) error {
	// get type of the interface
	itype := GetType[T]()

	// lock the container
	store.Lck.Lock()
	defer store.Lck.Unlock()

	// check if type is exist in the container
	if _, ok := store.instances[itype]; !ok {
		store.instances[itype] = make(map[InstanceName]interface{})
	}

	// check if instance is exist
	if _, ok := store.instances[itype][name]; ok {
		return ErrInstanceExist
	}

	// add instance to the container
	store.instances[itype][name] = instance

	return nil
}

// ReplaceInstance replaces instance in the container
func ReplaceInstance[T any](name InstanceName, instance T) {
	// get type of the interface
	itype := GetType[T]()

	// lock the container
	store.Lck.Lock()
	defer store.Lck.Unlock()

	// check if type is exist in the container
	if _, ok := store.instances[itype]; !ok {
		store.instances[itype] = make(map[InstanceName]interface{})
	}

	// add instance to the container
	store.instances[itype][name] = instance
}

// DeleteInstance deletes instance from the container
func DeleteInstance[T any](name InstanceName) {
	// get type of the interface
	itype := GetType[T]()

	// lock the container
	store.Lck.Lock()
	defer store.Lck.Unlock()

	// check if type is exist in the container
	if _, ok := store.instances[itype]; !ok {
		return
	}

	// delete instance from the container
	delete(store.instances[itype], name)
}

// GetType returns the T full name (package name + type name)
func GetType[T any]() string {
	instanceType := reflect.TypeOf((*T)(nil)).Elem()
	itype := instanceType.PkgPath() + "." + instanceType.Name()

	return itype
}
