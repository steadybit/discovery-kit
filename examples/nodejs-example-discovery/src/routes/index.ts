// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { DiscoveryList } from '@steadybit/discovery-kit-api';
import express from 'express';

export const router = express.Router();

router.get('/', (_, res) => {
	const response: DiscoveryList = {
		discoveries: [
			{ method: 'GET', path: '/discoveries/cats' },
			{ method: 'GET', path: '/discoveries/dogs' },
		],
		targetTypes: [
			{ method: 'GET', path: '/targettypes/cat' },
			{ method: 'GET', path: '/targettypes/dog' },
		],
		targetAttributes: [{ method: 'GET', path: '/targetattributes/pets' }],
	};
	res.json(response);
});
