// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { validate } from './util';
import schema from '../schema/discoverResponse.json';

describe('discoverResponse', () => {
	it('must support empty list', () => {
		expect(
			validate(schema, {
				targets: [],
			}).valid
		).toEqual(true);
	});

	it('must support minimum required fields', () => {
		expect(
			validate(schema, {
				targets: [{ id: 'test', targetType: 'cat', attributes: { 'cat.name': ['Garfield'] } }],
			}).valid
		).toEqual(true);
	});

	it('must report missing fields', () => {
		expect(
			validate(schema, {
				id: 'test-discovery',
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
