package roaringidx

import (
	"github.com/echoface/be_indexer"
	"github.com/echoface/be_indexer/util"
)

type (
	BEContainer interface {
		Meta() *FieldMeta

		AddWildcard(id ConjunctionID)

		Retrieve(values be_indexer.Values, inout *PostingList) error
	}

	BEContainerBuilder interface {
		EncodeWildcard(id ConjunctionID) // equal to: EncodeExpr(id ConjunctionID, nil)

		EncodeExpr(id ConjunctionID, expr *be_indexer.BooleanExpr) error

		BuildBEContainer() (BEContainer, error)
	}

	// DefaultBEContainer a common value based inverted index bitmap container
	DefaultBEContainer struct {
		meta *FieldMeta

		wc PostingList

		inc map[BEValue]PostingList

		exc map[BEValue]PostingList
	}
)

func NewDefaultBEContainer(meta *FieldMeta) *DefaultBEContainer {
	util.PanicIf(meta.Parser == nil, "default container must need parser")

	return &DefaultBEContainer{
		meta: meta,
		wc:   NewPostingList(),
		inc:  map[BEValue]PostingList{},
		exc:  map[BEValue]PostingList{},
	}
}

func (c *DefaultBEContainer) Meta() *FieldMeta {
	return c.meta
}

func (c *DefaultBEContainer) AddWildcard(id ConjunctionID) {
	c.wc.Add(uint64(id))
}

func (c *DefaultBEContainer) AddInclude(value BEValue, id ConjunctionID) {
	pl, ok := c.inc[value]
	if !ok {
		pl = NewPostingList()
		c.inc[value] = pl
	}
	pl.Add(uint64(id))
}

func (c *DefaultBEContainer) AddExclude(value BEValue, id ConjunctionID) {
	pl, ok := c.exc[value]
	if !ok {
		pl = NewPostingList()
		c.exc[value] = pl
	}
	pl.Add(uint64(id))
	// c.AddWildcard(id)
}

func (c *DefaultBEContainer) Retrieve(values be_indexer.Values, inout *PostingList) error {
	inout.Or(c.wc.Bitmap)

	if util.NilInterface(values) {
		return nil
	}

	ids, err := c.meta.Parser.ParseAssign(values)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if incPl, ok := c.inc[BEValue(id)]; ok {
			inout.Or(incPl.Bitmap)
		}
	}
	for _, id := range ids {
		if excPl, ok := c.exc[BEValue(id)]; ok {
			inout.AndNot(excPl.Bitmap)
		}
	}
	return nil
}

func (c *DefaultBEContainer) EncodeWildcard(id ConjunctionID) {
	c.AddWildcard(id)
}

func (c *DefaultBEContainer) EncodeExpr(id ConjunctionID, expr *be_indexer.BooleanExpr) error {
	if expr == nil {
		return nil
		// c.EncodeWildcard(id)
	}
	util.PanicIf(expr.Operator != be_indexer.ValueOptEQ, "default container support EQ operator only")

	valueIDs, err := c.meta.Parser.ParseValue(expr.Value)
	if err != nil {
		return err
	}
	for _, value := range valueIDs {
		if expr.Incl {
			c.AddInclude(BEValue(value), id)
		} else {
			c.AddExclude(BEValue(value), id)
		}
	}
	return nil
}

func (c *DefaultBEContainer) BuildBEContainer() (BEContainer, error) {
	//for _, v := range builder.container.inc {
	//	v.RunOptimize()
	//}
	//for _, v := range builder.container.exc {
	//	v.RunOptimize()
	//}
	//builder.container.wc.RunOptimize()
	return c, nil
}
