package common

// topologicalSort performs a topological sort on a set of items given their dependencies.
func TopologicalSort[T comparable](items []T, deps map[T][]T) []T {
	// Create a map to track the state of each item: 0 = unvisited, 1 = visiting, 2 = visited
	state := make(map[T]int)
	// Create a slice to store the sorted order
	var sorted []T
	// Variable to detect cycles
	hasCycle := false

	// Define the DFS function
	var visit func(T)
	visit = func(item T) {
		// If we detect a cycle, we can stop early
		if hasCycle {
			return
		}

		switch state[item] {
		case 1: // Currently visiting, so we have a cycle
			hasCycle = true
			return
		case 2: // Already visited
			return
		}

		// Mark the item as visiting
		state[item] = 1

		// Visit all dependencies (children) of the current item
		for _, dep := range deps[item] {
			visit(dep)
		}

		// Mark the item as visited
		state[item] = 2
		// Add the item to the sorted list
		sorted = append(sorted, item)
	}

	// Visit all items
	for _, item := range items {
		if state[item] == 0 {
			visit(item)
		}
	}

	// If we detected a cycle, return an empty slice
	if hasCycle {
		return []T{}
	}

	// The sorted list needs to be reversed since we used a stack-like approach
	for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
		sorted[i], sorted[j] = sorted[j], sorted[i]
	}

	return sorted
}
