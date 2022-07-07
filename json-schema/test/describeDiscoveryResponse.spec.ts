// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { validate } from './util';
import schema from '../schema/describeDiscoveryResponse.json';

describe('describeDiscoveryResponse', () => {
	it('must support minimum required fields', () => {
		expect(
			validate(schema, {
				id: 'test-discovery',
				discover: {
					method: 'GET',
					path: '/discover',
					callInterval: '10s',
				},
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
