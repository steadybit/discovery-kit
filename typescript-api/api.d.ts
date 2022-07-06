// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

export type Method = 'GET' | 'POST' | 'PUT' | 'DELETE';

export interface HttpEndpointRef<ALLOWED_METHODS extends Method> {
	method?: ALLOWED_METHODS;
	path: string;
}

export interface HttpEndpointRefWithCallInternval<ALLOWED_METHODS extends Method>
	extends HttpEndpointRef<ALLOWED_METHODS> {
	/**
	 * The duration, e.g., `100ms` or `1s`.
	 */
	callInterval?: string;
}

export type IndexResponse = SuccessfulIndexResponse | Problem;

export interface SuccessfulIndexResponse {
	discoveries?: HttpEndpointRef<'GET'>[];
	targetTypes?: HttpEndpointRef<'GET'>[];
	targetAttributes?: HttpEndpointRef<'GET'>[];
}

export type DescribeDiscoveryResponse = SuccessfulDescribeDiscoveryResponse | Problem;

export interface SuccessfulDescribeDiscoveryResponse {
	id: string;
	discover: HttpEndpointRefWithCallInternval<'GET'>;
}

export type DiscoverResponse = SuccessfulDiscoverResponse | Problem;

export interface SuccessfulDiscoverResponse {
	targets: Target[];
}

export type DescribeTargetTypeResponse = SuccessfulDescribeTargetTypeResponse | Problem;

export interface SuccessfulDescribeTargetTypeResponse {
	id: string;
	version: string;
	label: PuralLabel;
	icon: string;
	table: Table;
}

export type DescribeTargetAttributeResponse = SuccessfulDescribeTargetAttributeResponse | Problem;

export interface SuccessfulDescribeTargetAttributeResponse {
	attributes: TargetAttributeDescription[];
}

export interface Target {
	id: string;
	label?: string;
	targetType: string;
	attributes: Record<string, string[]>;
}

export interface PuralLabel {
	one: string;
	other: string;
}

export interface Table {
	columns: Column[];
	orderBy?: Order[];
}

export interface Column {
	attribute: string;
	fallbackAttributes?: string[];
}

export interface Order {
	attribute: string;
	direction: 'ASC' | 'DESC';
}

export interface TargetAttributeDescription {
	attribute: string;
	label: PuralLabel;
}

export interface Problem {
	type?: string;
	title: string;
	detail?: string;
	instance?: string;
}
