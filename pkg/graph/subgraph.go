package graph

import (
	"math"
	"sort"
)

// ComputeSubgraphs detects connected components in the dependency graph and computes scores
func (g *DependencyGraph) ComputeSubgraphs() {
	if len(g.Nodes) == 0 {
		return
	}

	// Build undirected adjacency list (treat edges as bidirectional for connectivity)
	adjacency := make(map[string][]string)
	for nodeID := range g.Nodes {
		adjacency[nodeID] = make([]string, 0)
	}

	// Add forward edges
	for source, targets := range g.Edges {
		adjacency[source] = append(adjacency[source], targets...)
		// Add reverse edges for connectivity detection
		for _, target := range targets {
			if _, exists := adjacency[target]; exists {
				adjacency[target] = append(adjacency[target], source)
			}
		}
	}

	// Find connected components using DFS
	visited := make(map[string]bool)
	subgraphID := 0
	g.Subgraphs = make([]Subgraph, 0)

	for nodeID := range g.Nodes {
		if !visited[nodeID] {
			// Start new subgraph
			component := make([]string, 0)
			dfs(nodeID, adjacency, visited, &component)

			// Create subgraph
			subgraph := Subgraph{
				ID:      subgraphID,
				NodeIDs: component,
			}

			// Count edges within this subgraph
			nodeSet := make(map[string]bool)
			for _, nid := range component {
				nodeSet[nid] = true
			}

			edgeCount := 0
			for _, nid := range component {
				if targets, exists := g.Edges[nid]; exists {
					for _, target := range targets {
						if nodeSet[target] {
							edgeCount++
						}
					}
				}
			}
			subgraph.EdgeCount = edgeCount

			// Compute score
			subgraph.Score = computeSubgraphScore(len(component), edgeCount)

			// Assign subgraph metadata to all nodes in this component
			for _, nid := range component {
				if node, exists := g.Nodes[nid]; exists {
					node.SubgraphID = subgraphID
					node.SubgraphScore = subgraph.Score
				}
			}

			g.Subgraphs = append(g.Subgraphs, subgraph)
			subgraphID++
		}
	}

	// Sort subgraphs by score (descending) for easier identification of important clusters
	sort.Slice(g.Subgraphs, func(i, j int) bool {
		return g.Subgraphs[i].Score > g.Subgraphs[j].Score
	})

	// Reassign subgraph IDs after sorting
	for i := range g.Subgraphs {
		g.Subgraphs[i].ID = i
		for _, nodeID := range g.Subgraphs[i].NodeIDs {
			if node, exists := g.Nodes[nodeID]; exists {
				node.SubgraphID = i
			}
		}
	}
}

// dfs performs depth-first search to find all nodes in a connected component
func dfs(nodeID string, adjacency map[string][]string, visited map[string]bool, component *[]string) {
	visited[nodeID] = true
	*component = append(*component, nodeID)

	for _, neighbor := range adjacency[nodeID] {
		if !visited[neighbor] {
			dfs(neighbor, adjacency, visited, component)
		}
	}
}

// computeSubgraphScore calculates a score for a subgraph based on its properties
// Score formula: nodeCount * log2(nodeCount + 1) + edgeCount * 2
// This gives higher weight to:
// - Larger subgraphs (more nodes)
// - Better connected subgraphs (more edges)
// - Uses logarithmic scaling for nodes to prevent huge subgraphs from dominating
func computeSubgraphScore(nodeCount, edgeCount int) float64 {
	if nodeCount == 0 {
		return 0.0
	}

	// Base score from node count with logarithmic scaling
	nodeScore := float64(nodeCount) * math.Log2(float64(nodeCount+1))

	// Edge score (edges indicate stronger connectivity)
	edgeScore := float64(edgeCount) * 2.0

	// Density bonus: reward graphs that are well-connected relative to their size
	// Maximum possible edges in a directed graph: n * (n - 1)
	maxPossibleEdges := nodeCount * (nodeCount - 1)
	if maxPossibleEdges > 0 {
		density := float64(edgeCount) / float64(maxPossibleEdges)
		densityBonus := density * float64(nodeCount) * 5.0
		return nodeScore + edgeScore + densityBonus
	}

	return nodeScore + edgeScore
}

// GetSubgraphByID returns a subgraph by its ID
func (g *DependencyGraph) GetSubgraphByID(id int) *Subgraph {
	for i := range g.Subgraphs {
		if g.Subgraphs[i].ID == id {
			return &g.Subgraphs[i]
		}
	}
	return nil
}

// GetNodeSubgraph returns the subgraph that contains the given node
func (g *DependencyGraph) GetNodeSubgraph(nodeID string) *Subgraph {
	if node, exists := g.Nodes[nodeID]; exists {
		return g.GetSubgraphByID(node.SubgraphID)
	}
	return nil
}

// GetLargestSubgraph returns the subgraph with the highest score
func (g *DependencyGraph) GetLargestSubgraph() *Subgraph {
	if len(g.Subgraphs) == 0 {
		return nil
	}
	// Subgraphs are already sorted by score, so first one is largest
	return &g.Subgraphs[0]
}
