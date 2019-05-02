package yamled

import (
	"fmt"
	"strconv"
	"strings"
)

type Step interface{}

type Path []Step

func (p Path) Parent() Path {
	if len(p) < 1 {
		return nil
	}

	return p[0 : len(p)-1]
}

func (p Path) Tail() Step {
	if len(p) == 0 {
		return nil
	}

	return p[len(p)-1]
}

func (p Path) String() string {
	items := []string{}

	for _, item := range p {
		if s, ok := item.(string); ok {
			if strings.Contains(s, ".") {
				s = fmt.Sprintf(`"%s"`, s)
			}

			items = append(items, s)
		} else if i, ok := item.(int); ok {
			items = append(items, strconv.Itoa(i))
		}
	}

	return strings.Join(items, ".")
}
