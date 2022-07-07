// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { validate } from './util';
import schema from '../schema/describeTargetTypeResponse.json';

describe('describeTargetTypeResponse', () => {
	it('must support minimum required fields', () => {
		expect(
			validate(schema, {
				id: 'cat',
				version: '0.0.1',
				label: { one: 'cat', other: 'cats' },
				icon: 'data::',
				table: {
					columns: [
						{
							attribute: 'cat.name',
						},
						{ attribute: 'cat.age' },
					],
				},
			}).valid
		).toEqual(true);
	});

	it('must support additional fields', () => {
		expect(
			validate(schema, {
				id: 'cat',
				version: '0.0.1',
				label: { one: 'cat', other: 'cats' },
				icon: 'data::',
				table: {
					columns: [
						{
							attribute: 'cat.name',
							fallbackAttributes: ['cat.secondName'],
						},
						{ attribute: 'cat.age' },
					],
					orderBy: [{ attribute: 'cat.name', direction: 'ASC' }],
				},
			}).valid
		).toEqual(true);
	});

	it('must report missing fields', () => {
		expect(
			validate(schema, {
				id: 'cat',
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
