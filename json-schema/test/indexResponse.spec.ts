// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { validate } from './util';
import schema from '../schema/indexResponse.json';

describe('indexResponse', () => {
	it('must support empty lists', () => {
		expect(
			validate(schema, {
				discoveries: [],
				targetTypes: [],
				targetAttributes: [],
			}).valid
		).toEqual(true);
	});

	it('must support empty object', () => {
		expect(validate(schema, {}).valid).toEqual(true);
	});

	it('must support single discovery reference', () => {
		expect(
			validate(schema, {
				discoveries: [
					{
						method: 'GET',
						path: '/list',
					},
				],
			}).valid
		).toEqual(true);
	});

	it('must support single targetType reference', () => {
		expect(
			validate(schema, {
				targetTypes: [
					{
						method: 'GET',
						path: '/list',
					},
				],
			}).valid
		).toEqual(true);
	});

	it('must support single targetAttribute reference', () => {
		expect(
			validate(schema, {
				targetAttributes: [
					{
						method: 'GET',
						path: '/list',
					},
				],
			}).valid
		).toEqual(true);
	});

	it('must identify invalid references', () => {
		expect(
			validate(schema, {
				discoveries: [
					{
						method: 'POST',
						path: 'non-absolute',
					},
				],
			}).valid
		).toEqual(false);
	});
});
