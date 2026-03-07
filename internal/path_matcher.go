package internal

import (
	"fmt"
	"strings"
)

type Node struct {
	children      map[string]*Node
	wildcardChild *Node
	rules         map[string]*RuleSet
}

type ServiceConfig struct {
	algorithm string
	root      *Node
}

type PathMatcher struct {
	services map[string]*ServiceConfig
}

func NewPathMatcher() *PathMatcher {
	return &PathMatcher{
		services: make(map[string]*ServiceConfig),
	}
}

func (p *PathMatcher) Insert(path *Path, service string, algorithm string) {
	if service == "" {
		service = "default"
	}

	// if service is already added then keep service same else insert service and add config
	config, exists := p.services[service]

	if !exists {
		config = &ServiceConfig{algorithm: algorithm, root: NewNode()}
		p.services[service] = config
	}

	current := config.root
	parts := strings.Split(strings.Trim(path.Path, "/"), "/")

	for _, part := range parts {
		if part == "" {
			continue
		}

		if part == "*" || part[0] == ':' {
			//this is a wildcard
			if current.wildcardChild == nil {
				current.wildcardChild = NewNode()
			}
			current = current.wildcardChild
		} else {
			if _, ok := current.children[part]; !ok {
				current.children[part] = NewNode()
			}
			current = current.children[part]
		}
	}
	if current.rules == nil {
		current.rules = make(map[string]*RuleSet)
	}
	current.rules[path.Method] = path.Rules
}

func (p *PathMatcher) Search(path string, service string, method string) (*RuleSet, string, error) {
	if service == "" {
		service = "default"
	}

	// if service is already added then keep service same else insert service and add config
	config, exists := p.services[service]

	if !exists {
		return &RuleSet{}, "", fmt.Errorf("service %s not found", service)
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	current := config.root

	for _, part := range parts {
		if next, ok := current.children[part]; ok {
			current = next
		} else if current.wildcardChild != nil {
			current = current.wildcardChild
		} else {
			return &RuleSet{}, "", fmt.Errorf("no rules available for path %s ", path)
		}
	}

	rules, exists := current.rules[method]
	if !exists {
		return &RuleSet{}, "", fmt.Errorf("no rules available for method, path %s %s", method, path)
	}
	return rules, config.algorithm, nil
}

func NewNode() *Node {
	return &Node{
		children:      make(map[string]*Node),
		wildcardChild: nil,
		rules:         make(map[string]*RuleSet),
	}
}
