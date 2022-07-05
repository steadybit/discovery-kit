// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { DescribeDiscoveryResponse, DiscoverResponse } from '@steadybit/discovery-api';
import express from 'express';

export const router = express.Router();

router.get('/discoveries/cats', (_, res) => {
	const response: DescribeDiscoveryResponse = {
		id: 'cats-discovery',
		discover: {
			path: '/discoveries/cats/discover',
			callInterval: '10s',
		},
	};
	res.json(response);
});

router.get('/discoveries/cats/discover', (req, res) => {
	console.log('Got discover request');
	const response: DiscoverResponse = {
		targets: [
			{
				name: 'garfield',
				targetType: 'cat',
				attributes: {
					'pet.name': ['Garfield'],
					'pet.age': ['42'],
					'pet.owner': ['Daniel'],
				},
			},
			{
				name: 'kitty',
				targetType: 'cat',
				attributes: {
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
	const response: DescribeDiscoveryResponse = {
		id: 'dogs-discovery',
		discover: {
			path: '/discoveries/dogs/discover',
			callInterval: '10s',
		},
	};
	res.json(response);
});

router.get('/discoveries/dogs/discover', (req, res) => {
	console.log('Got discover request');
	const response: DiscoverResponse = {
		targets: [
			{
				name: 'emma',
				targetType: 'dog',
				attributes: {
					'pet.name': ['Emma'],
					'pet.age': ['2'],
					'pet.owner': ['Daniel'],
				},
			},
			{
				name: 'lassie',
				targetType: 'dog',
				attributes: {
					'pet.name': ['Lassie'],
					'pet.age': ['7'],
					'pet.owner': ['Johannes'],
				},
			},
		],
	};
	res.json(response);
});
