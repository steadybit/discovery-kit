// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { validate } from './util';
import schema from '../schema/describeTargetAttributeResponse.json';

describe('describeTargetAttributeResponse', () => {
	it('must support minimum required fields', () => {
		expect(
			validate(schema, {
				attributes: [{ attribute: 'cat.name', label: { one: 'cat name', other: 'cat names' } }],
			}).valid
		).toEqual(true);
	});

	it('must report missing fields', () => {
		expect(
			validate(schema, {
				attributes: [{ attribute: 'cat.name' }],
			}).valid
		).toEqual(false);
	});

	it('must support rfc 7807 problems', () => {
		expect(
			validate(schema, {
				title: 'Something went wrong',
				details: 'Terrible things happens',
			}).valid
		).toEqual(true);
	});
});
