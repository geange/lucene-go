package util

import (
	"sync"
)

// An AttributeSource contains a list of different AttributeImpls, and methods to add and get them.
// There can only be a single instance of an attribute in the same AttributeSource instance. This is
// ensured by passing in the actual type of the Attribute (Class<Attribute>) to the addAttribute(Class),
// which then checks if an instance of that type is already present. If yes, it returns the instance,
// otherwise it creates a new instance and returns it.
type AttributeSource struct {
	attributes sync.Map
	list       []AttributeImpl
}

func (a *AttributeSource) Get(name string) (AttributeImpl, bool) {

	v, ok := a.attributes.Load(name)
	return v.(AttributeImpl), ok
}

func (a *AttributeSource) Add(item AttributeImpl) {
	isDuplicate := true
	for _, name := range item.Interfaces() {
		if _, ok := a.attributes.Load(name); !ok {
			isDuplicate = false
			a.attributes.Store(name, item)
		}
	}

	if !isDuplicate {
		a.list = append(a.list, item)
	}
}

func (a *AttributeSource) Clear() {

}
