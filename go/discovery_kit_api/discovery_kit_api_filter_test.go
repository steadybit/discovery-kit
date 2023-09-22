/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package discovery_kit_api

import (
	"reflect"
	"testing"
)

var targets = []Target{
	{
		Id:         "id1",
		TargetType: "type1",
		Label:      "label1",
		Attributes: map[string][]string{
			"key1":                      {"value1"},
			"key2":                      {"value2"},
			"aws.ec2.label.Environment": {"production"},
		},
	},
	{
		Id:         "id2",
		TargetType: "type2",
		Label:      "label2",
		Attributes: map[string][]string{
			"key1":                      {"value1"},
			"key2":                      {"value2"},
			"aws.account":               {"123456789012"},
			"aws.region":                {"eu-central-1"},
			"aws.ec2.instance-id":       {"i-1234567890abcdef0"},
			"aws.ec2.instance-type":     {"t2.micro"},
			"aws.ec2.tag.Name":          {"my-instance"},
			"aws.ec2.tag.Environment":   {"production"},
			"aws.ec2.label.Name":        {"my-instance"},
			"aws.ec2.label.Environment": {"production"},
			"aws.ec2.label":             {"label"},
		},
	},
}

func TestApplyDenyList(t *testing.T) {
	type args struct {
		targets  []Target
		denyList []string
	}
	tests := []struct {
		name string
		args args
		want []Target
	}{
		{
			name: "empty deny list",
			args: args{
				targets:  targets,
				denyList: []string{},
			},
			want: targets,
		}, {
			name: "nil deny list",
			args: args{
				targets:  targets,
				denyList: nil,
			},
			want: targets,
		},
		{
			name: "deny list with one entry",
			args: args{
				targets:  targets,
				denyList: []string{"key1"},
			},
			want: []Target{
				{
					Id:         "id1",
					TargetType: "type1",
					Label:      "label1",
					Attributes: map[string][]string{
						"key2":                      {"value2"},
						"aws.ec2.label.Environment": {"production"},
					},
				},
				{
					Id:         "id2",
					TargetType: "type2",
					Label:      "label2",
					Attributes: map[string][]string{
						"key2":                      {"value2"},
						"aws.account":               {"123456789012"},
						"aws.region":                {"eu-central-1"},
						"aws.ec2.instance-id":       {"i-1234567890abcdef0"},
						"aws.ec2.instance-type":     {"t2.micro"},
						"aws.ec2.tag.Name":          {"my-instance"},
						"aws.ec2.tag.Environment":   {"production"},
						"aws.ec2.label.Name":        {"my-instance"},
						"aws.ec2.label.Environment": {"production"},
						"aws.ec2.label":             {"label"},
					},
				},
			},
		},
		{
			name: "deny list with one entry with wildcard",
			args: args{
				targets:  targets,
				denyList: []string{"key*"},
			},
			want: []Target{
				{
					Id:         "id1",
					TargetType: "type1",
					Label:      "label1",
					Attributes: map[string][]string{
						"aws.ec2.label.Environment": {"production"},
					},
				},
				{
					Id:         "id2",
					TargetType: "type2",
					Label:      "label2",
					Attributes: map[string][]string{
						"aws.account":               {"123456789012"},
						"aws.region":                {"eu-central-1"},
						"aws.ec2.instance-id":       {"i-1234567890abcdef0"},
						"aws.ec2.instance-type":     {"t2.micro"},
						"aws.ec2.tag.Name":          {"my-instance"},
						"aws.ec2.tag.Environment":   {"production"},
						"aws.ec2.label.Name":        {"my-instance"},
						"aws.ec2.label.Environment": {"production"},
						"aws.ec2.label":             {"label"},
					},
				},
			},
		},
		{
			name: "deny list with with wildcards",
			args: args{
				targets:  targets,
				denyList: []string{"key.*", "aws.ec2.*"},
			},
			want: []Target{
				{
					Id:         "id1",
					TargetType: "type1",
					Label:      "label1",
					Attributes: map[string][]string{
						"key1": {"value1"},
						"key2": {"value2"},
					},
				},
				{
					Id:         "id2",
					TargetType: "type2",
					Label:      "label2",
					Attributes: map[string][]string{
						"key1":        {"value1"},
						"key2":        {"value2"},
						"aws.account": {"123456789012"},
						"aws.region":  {"eu-central-1"},
					},
				},
			},
		},
		{
			name: "deny list with with longer keys and wildcards",
			args: args{
				targets:  targets,
				denyList: []string{"aws.ec2.label.*"},
			},
			want: []Target{
				{
					Id:         "id1",
					TargetType: "type1",
					Label:      "label1",
					Attributes: map[string][]string{
						"key1": {"value1"},
						"key2": {"value2"},
					},
				},
				{
					Id:         "id2",
					TargetType: "type2",
					Label:      "label2",
					Attributes: map[string][]string{
						"key1":                    {"value1"},
						"key2":                    {"value2"},
						"aws.account":             {"123456789012"},
						"aws.region":              {"eu-central-1"},
						"aws.ec2.instance-id":     {"i-1234567890abcdef0"},
						"aws.ec2.instance-type":   {"t2.micro"},
						"aws.ec2.tag.Name":        {"my-instance"},
						"aws.ec2.tag.Environment": {"production"},
						"aws.ec2.label":           {"label"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ApplyAttributeExcludes(tt.args.targets, tt.args.denyList); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplyAttributeExcludes() = %v, want %v", got, tt.want)
			}
		})
	}
}
