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
					'cat.name': ['Garfield'],
					'cat.age': ['42'],
					'cat.owner': ['Daniel'],
				},
			},
			{
				name: 'kitty',
				targetType: 'cat',
				attributes: {
					'cat.name': ['Kitty'],
					'cat.age': ['0'],
					'cat.owner': ['Ben'],
				},
			},
		],
	};
	res.json(response);
});
