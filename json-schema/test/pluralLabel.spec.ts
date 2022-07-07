// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { validate } from './util';
import schema from '../schema/pluralLabel.json';

describe('pluralLabel', () => {
	it('must support required attributes', () => {
		expect(
			validate(schema, {
				one: 'cat',
				other: 'cats',
			}).valid
		).toEqual(true);
	});

	it('must identify invalid labels', () => {
		expect(validate(schema, { one: null, other: 'cats' }).valid).toEqual(false);
		expect(validate(schema, { one: 'cat' }).valid).toEqual(false);
		expect(validate(schema, { other: 'cats' }).valid).toEqual(false);
	});
});
