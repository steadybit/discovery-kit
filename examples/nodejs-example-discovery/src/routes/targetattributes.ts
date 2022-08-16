// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { AttributeDescriptions } from '@steadybit/discovery-kit-api';
import express from 'express';

export const router = express.Router();

router.get('/targetattributes/pets', (_, res) => {
	const response: AttributeDescriptions = {
		attributes: [
			{ attribute: 'pet.name', label: { one: 'pet name', other: 'pet names' } },
			{ attribute: 'pet.age', label: { one: 'pet age', other: 'pet ages' } },
			{ attribute: 'pet.owner', label: { one: 'pet owner', other: 'pet owner' } },
		],
	};
	res.json(response);
});
