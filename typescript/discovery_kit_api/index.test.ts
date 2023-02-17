// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

import {
	AttributeDescriptions, DiscoveredTargets,
	DiscoveryDescription,
	DiscoveryKitError,
	DiscoveryList, TargetDescription

} from "./index";

export const discoveryList: DiscoveryList = {
	discoveries: [
		{
			method: "GET",
			path: "/action"
		}
	],

	targetTypes: [
		{
			method: "GET",
			path: "/action"
		}
	],

	targetAttributes: [
		{
			method: "GET",
			path: "/action"
		}
	]
};

export const discoveryDescription: DiscoveryDescription = {
	discover: {
		method: "GET",
		path: "/",
		callInterval: "5m"
	},
	id: "42",
	restrictTo: "LEADER"
};

export const attributeDescriptions: AttributeDescriptions = {
	attributes: [
		{
			attribute: "k8s.deployment",
			label: {
				one: "Kubernetes deployment",
				other: "Kubernetes deployments"
			}
		}
	]
};

export const discoveredTargets: DiscoveredTargets = {
	targets: [
		{
			attributes: {
				foo: ["bar"]
			},
			id: "i",
			label: "l",
			targetType: "t"
		}
	]
};

export const targetDescription: TargetDescription = {
	category: 'basic',
	icon: 'data:...',
	id: 'id',
	version: '1.0.0',
	label: {
		one: 'one',
		other: 'other'
	},
	table: {
		columns: [
			{
				attribute: 'attr',
				fallbackAttributes: ['a', 'b']
			}
		],
		orderBy: [
			{
				attribute: 'attr',
				direction: 'DESC'
			}
		]
	},
	enrichmentRules: [
		{
			src: {
				type: 'k8s.deployment',
				selector: {
					"container.id": "${dest.container.id}",
				}
			},
			dest: {
				type: 'k8s.deployment',
				selector: {
					"container.id": "${dest.container.id}",
				}
			},
			attributes: [
				{
					aggregationType: "all",
					name: 'container.name'
				},
				{
					aggregationType: "any",
					name: 'container.name'
				}
			]
		}
	]
};

export const discoveryKitError: DiscoveryKitError = {
	detail: "d",
	instance: "i",
	title: "t",
	type: "t"
};
