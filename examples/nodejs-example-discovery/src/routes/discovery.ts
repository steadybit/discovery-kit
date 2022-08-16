// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { DiscoveryDescription, DiscoveredTargets } from '@steadybit/discovery-kit-api';
import express from 'express';

export const router = express.Router();

router.get('/discoveries/cats', (_, res) => {
	const response: DiscoveryDescription = {
		id: 'cats-discovery',
		discover: {
			method: 'GET',
			path: '/discoveries/cats/discover',
			callInterval: '10s',
		},
		restrictTo: 'LEADER',
	};
	res.json(response);
});

router.get('/discoveries/cats/discover', (req, res) => {
	console.log('Got discover request');
	const response: DiscoveredTargets = {
		targets: [
			{
				id: 'garfield',
				label: 'Garfield',
				targetType: 'cat',
				attributes: {
					'steadybit.label': ['Garfield'],
					'pet.name': ['Garfield'],
					'pet.age': ['42'],
					'pet.owner': ['Daniel'],
				},
			},
			{
				id: 'kitty',
				label: 'Kitty',
				targetType: 'cat',
				attributes: {
					'steadybit.label': ['Kitty'],
					'pet.name': ['Kitty'],
					'pet.age': ['0'],
					'pet.owner': ['Ben'],
				},
			},
		],
	};
	res.json(response);
});

router.get('/discoveries/dogs', (_, res) => {
	const response: DiscoveryDescription = {
		id: 'dogs-discovery',
		discover: {
			method: 'GET',
			path: '/discoveries/dogs/discover',
			callInterval: '10s',
		},
		restrictTo: 'LEADER',
	};
	res.json(response);
});

router.get('/discoveries/dogs/discover', (req, res) => {
	console.log('Got discover request');
	const response: DiscoveredTargets = {
		targets: [
			{
				id: 'emma',
				label: 'Emma',
				targetType: 'dog',
				attributes: {
					'steadybit.label': ['Emma'],
					'pet.name': ['Emma'],
					'pet.age': ['2'],
					'pet.owner': ['Daniel'],
				},
			},
			{
				id: 'lassie',
				label: 'Lassie',
				targetType: 'dog',
				attributes: {
					'steadybit.label': ['Lassie'],
					'pet.name': ['Lassie'],
					'pet.age': ['7'],
					'pet.owner': ['Johannes'],
				},
			},
		],
	};
	res.json(response);
});
