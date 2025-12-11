package main

import (
	_ "embed"
	"fmt"
	"regexp"

	"go.yaml.in/yaml/v3"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

//go:embed assets/rbac.yaml
var rbacs string

var yamlDocSplitter = regexp.MustCompile(`(?m)^---$`)

func getRules() (rules []rbacv1.PolicyRule, clusterRules []rbacv1.PolicyRule, err error) {
	roles, clusterRoles, err := getRolesObjects()
	if err != nil {
		return nil, nil, err
	}

	rules = getRulesFromRoles(roles)
	clusterRules = getRulesFromClusterRole(clusterRoles)

	return rules, clusterRules, nil
}

func getRulesFromRoles(roles []rbacv1.Role) []rbacv1.PolicyRule {
	var rules []rbacv1.PolicyRule
	for _, role := range roles {
		rules = append(rules, role.Rules...)
	}

	return rules
}

func getRulesFromClusterRole(clusterRoles []rbacv1.ClusterRole) []rbacv1.PolicyRule {
	var rules []rbacv1.PolicyRule
	for _, clusterRole := range clusterRoles {
		rules = append(rules, clusterRole.Rules...)
	}

	return rules
}

func getRolesObjects() ([]rbacv1.Role, []rbacv1.ClusterRole, error) {
	var (
		roles        []rbacv1.Role
		clusterRoles []rbacv1.ClusterRole
	)

	docs := yamlDocSplitter.Split(rbacs, -1)
	for _, doc := range docs {
		if len(doc) == 0 {
			continue
		}

		obj := unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(doc), &obj.Object)
		if err != nil {
			return nil, nil, fmt.Errorf("can't parse yaml object \n%s\n%w", doc, err)
		}

		switch obj.GetObjectKind().GroupVersionKind().GroupKind().String() {
		case "Role.rbac.authorization.k8s.io":
			role := rbacv1.Role{}
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &role); err != nil {
				return nil, nil, fmt.Errorf("can't convert unstructured object to Role; %w", err)
			}

			roles = append(roles, role)
		case "ClusterRole.rbac.authorization.k8s.io":
			clusterRole := rbacv1.ClusterRole{}
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &clusterRole); err != nil {
				return nil, nil, fmt.Errorf("can't convert unstructured object to ClusterRole; %w", err)
			}

			clusterRoles = append(clusterRoles, clusterRole)
		}
	}

	return roles, clusterRoles, nil
}
