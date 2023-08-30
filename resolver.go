package pgpkg

import (
	"fmt"
)

// sortPackages provides topological sort, which is used to determine the order
// in which packages should be installed, based on the dependency graph.
func (p *Project) sortPackages() ([]string, error) {
	visited := make(map[string]bool)
	var stack []string

	// Ensure that the pgpkg package itself is installed before anything else.
	if err := p.dfs(p.pkgs["github.com/pgpkg/pgpkg"], visited, &stack, make(map[string]bool)); err != nil {
		return nil, err
	}

	for _, pkg := range p.pkgs {
		pkgName := pkg.Name
		if !visited[pkgName] {
			if err := p.dfs(pkg, visited, &stack, make(map[string]bool)); err != nil {
				return nil, err
			}
		}
	}

	return stack, nil
}

// dfs is a recursive function that performs a depth-first search on the package dependency graph.
// It's used to determine the topological order of packages so that they are installed in the correct
// order.
//
// pkg: The current package being visited.
// visited: A map that keeps track of which package names have been visited.
// stack: A slice (acting as a stack) where nodes are added in post-order, meaning
// a node is added after all its dependencies have been visited.
//
// currentPath: A map that keeps track of the nodes in the current DFS path.
//
//	It's used to detect cycles.
func (p *Project) dfs(pkg *Package, visited map[string]bool, stack *[]string, currentPath map[string]bool) error {
	pkgName := pkg.Name
	visited[pkgName] = true
	currentPath[pkgName] = true

	for _, neighbor := range pkg.config.Uses {
		if currentPath[neighbor] {
			return fmt.Errorf("dependency cycle detected for package %s, uses %s", pkgName, neighbor)
		}
		if !visited[neighbor] {
			if err := p.dfs(p.pkgs[neighbor], visited, stack, currentPath); err != nil {
				return err
			}
		}
	}

	*stack = append(*stack, pkgName)
	delete(currentPath, pkgName)
	return nil
}
