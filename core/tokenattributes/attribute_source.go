package tokenattributes

import (
	"reflect"
	"sync"
)

// AttributeImpl Base class for Attributes that can be added to a AttributeSourceV2.
// Attributes are used to add data in a dynamic, yet types-safe way to a source of usually streamed objects,
// e. g. a org.apache.lucene.analysis.TokenStream.
type AttributeImpl interface {
	Interfaces() []string
	Clear() error
	End() error
	CopyTo(target AttributeImpl) error
	Clone() AttributeImpl
}

// An AttributeSourceV2 contains a list of different AttributeImpls, and methods to add and get them.
// There can only be a single instance of an attribute in the same AttributeSourceV2 instance. This is
// ensured by passing in the actual types of the Attribute (Class<Attribute>) to the addAttribute(Class),
// which then checks if an instance of that types is already present. If yes, it returns the instance,
// otherwise it creates a new instance and returns it.
type AttributeSourceV2 struct {
	attributes     sync.Map
	attributeImpls sync.Map
	list           []AttributeImpl
	factory        AttributeFactory
}

func NewAttributeSource() *AttributeSourceV2 {
	source := &AttributeSourceV2{
		attributes:     sync.Map{},
		attributeImpls: sync.Map{},
		list:           make([]AttributeImpl, 0),
		factory:        DEFAULT_ATTRIBUTE_FACTORY,
	}
	return source
}

func (a *AttributeSourceV2) Get(name string) (AttributeImpl, bool) {
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

func (a *AttributeSourceV2) Add(item AttributeImpl) {
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

func (a *AttributeSourceV2) Clear() error {
	for _, impl := range a.list {
		if err := impl.Clear(); err != nil {
			return err
		}
	}
	return nil
}
