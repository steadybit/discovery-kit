// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { IndexResponse } from '@steadybit/discovery-api';
import express from 'express';

export const router = express.Router();

router.get('/', (_, res) => {
	const response: IndexResponse = {
		discoveries: [{ path: '/discoveries/cats' }],
		targetTypes: [{ path: '/targettypes/cats' }],
		targetAttributes: [{ path: '/targetattributes/cats' }],
	};
	res.json(response);
});
