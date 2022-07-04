package tokenattributes

import (
	"reflect"
	"sync"
)

// AttributeImpl Base class for Attributes that can be added to a AttributeSource.
// Attributes are used to add data in a dynamic, yet types-safe way to a source of usually streamed objects,
// e. g. a org.apache.lucene.analysis.TokenStream.
type AttributeImpl interface {
	Interfaces() []string
	Clear() error
	End() error
	CopyTo(target AttributeImpl) error
	Clone() AttributeImpl
}

// An AttributeSource contains a list of different AttributeImpls, and methods to add and get them.
// There can only be a single instance of an attribute in the same AttributeSource instance. This is
// ensured by passing in the actual types of the Attribute (Class<Attribute>) to the addAttribute(Class),
// which then checks if an instance of that types is already present. If yes, it returns the instance,
// otherwise it creates a new instance and returns it.
type AttributeSource struct {
	attributes     sync.Map
	attributeImpls sync.Map
	list           []AttributeImpl
	factory        AttributeFactory
}

func NewAttributeSource() *AttributeSource {
	source := &AttributeSource{
		attributes:     sync.Map{},
		attributeImpls: sync.Map{},
		list:           make([]AttributeImpl, 0),
		factory:        DEFAULT_ATTRIBUTE_FACTORY,
	}
	return source
}

func (a *AttributeSource) Get(name string) (AttributeImpl, bool) {
	v, ok := a.attributes.Load(name)
	if !ok {
		impl, err := a.factory.CreateAttributeInstance(name)
		if err != nil {
			return nil, false
		}
		a.Add(impl)
		return impl, true
	}
	return v.(AttributeImpl), ok
}

func (a *AttributeSource) Add(item AttributeImpl) {
	for _, name := range item.Interfaces() {
		if _, ok := a.attributes.Load(name); !ok {
			a.attributes.Store(name, item)
		}
	}

	rType := reflect.TypeOf(item)
	if _, ok := a.attributeImpls.Load(rType); !ok {
		a.attributeImpls.Store(rType, item)
		a.list = append(a.list, item)
	}
}

func (a *AttributeSource) Clear() error {
	for _, impl := range a.list {
		if err := impl.Clear(); err != nil {
			return err
		}
	}
	return nil
}
