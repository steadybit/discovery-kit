// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { DescribeTargetAttributeResponse } from '@steadybit/discovery-api';
import express from 'express';

export const router = express.Router();

router.get('/targetattributes/cats', (_, res) => {
	const response: DescribeTargetAttributeResponse = {
		attributes: [
			{ attribute: 'cat.name', label: { one: 'cat name', other: 'cat names' } },
			{ attribute: 'cat.age', label: { one: 'cat age', other: 'cat ages' } },
			{ attribute: 'cat.owner', label: { one: 'cat owner', other: 'cat owner' } },
		],
	};
	res.json(response);
});
