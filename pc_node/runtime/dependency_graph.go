package runtime

import (
	"fmt"
)

type DependencyGraph struct {
	registry *ModuleRegistry
}

func NewDependencyGraph(registry *ModuleRegistry) *DependencyGraph {
	return &DependencyGraph{registry: registry}
}

// Build generates an ordered list of module names indicating startup order.
// Validates dependencies exist and checks for circular dependencies.
func (g *DependencyGraph) Build() ([]string, error) {
	modules := g.registry.All()
	adj := make(map[string][]string)

	// Check existence
	for _, info := range modules {
		name := info.Instance.Name()
		deps := info.Instance.Dependencies()
		for _, dep := range deps {
			if _, err := g.registry.Get(dep); err != nil {
				return nil, fmt.Errorf("module %s has unknown dependency: %s", name, dep)
			}
		}
		adj[name] = deps
	}

	// Kahn's algorithm for topological sort
	inDegree := make(map[string]int)
	for k := range adj {
		inDegree[k] = 0
	}
	for _, deps := range adj {
		for _, d := range deps {
			inDegree[d]++
		}
	}

	var queue []string
	for k, v := range inDegree {
		if v == 0 {
			queue = append(queue, k)
		}
	}

	var order []string
	for len(queue) > 0 {
		queue = queue[1:]
		// The original logic here was finding nodes with 0 incoming edges.
		// However, for startup, if A depends on B, B must start BEFORE A.
		// So A has an edge to B. B has incoming edge from A.
		// Therefore, we want to start nodes with 0 *outgoing* dependencies first?
		// No, if A depends on B, we can model it as directed edge B -> A (B unlocks A).
		// Let's rebuild the adjacency properly for Kahn's.
	}

	// Correct Kahn's implementation for startup order:
	// A depends on B.  (B -> A) edge.
	realAdj := make(map[string][]string)
	realInDegree := make(map[string]int)
	for _, info := range modules {
		realInDegree[info.Instance.Name()] = 0
	}

	for _, info := range modules {
		name := info.Instance.Name()
		for _, dep := range info.Instance.Dependencies() {
			// dep unlocks name
			realAdj[dep] = append(realAdj[dep], name)
			realInDegree[name]++
		}
	}

	queue = []string{}
	for k, v := range realInDegree {
		if v == 0 {
			queue = append(queue, k)
		}
	}

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		order = append(order, u)

		for _, v := range realAdj[u] {
			realInDegree[v]--
			if realInDegree[v] == 0 {
				queue = append(queue, v)
			}
		}
	}

	if len(order) != len(modules) {
		return nil, fmt.Errorf("circular dependency detected in module graph")
	}

	return order, nil
}
