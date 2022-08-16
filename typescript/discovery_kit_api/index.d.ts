// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { components } from './schemas';

export type DiscoveryList = components['schemas']['DiscoveryList'];
export type DiscoveryDescription = components['schemas']['DiscoveryDescription'];
export type DiscoveryKitError = components['schemas']['DiscoveryKitError'];
export type DescribingEndpointReference = components['schemas']['DescribingEndpointReference'];
export type DescribingEndpointReferenceWithCallInterval = components['schemas']['DescribingEndpointReferenceWithCallInterval'];
export type PluralLabel = components['schemas']['PluralLabel'];
export type AttributeDescription = components['schemas']['AttributeDescription'];
export type AttributeDescriptions = components['schemas']['AttributeDescriptions'];
export type Target = components['schemas']['Target'];
export type DiscoveredTargets = components['schemas']['DiscoveredTargets'];
export type OrderBy = components['schemas']['OrderBy'];
export type Column = components['schemas']['Column'];
export type Table = components['schemas']['Table'];
export type TargetDescription = components['schemas']['TargetDescription'];

export type DiscoveryListResponse = components['responses']['DiscoveryListResponse']['content']['application/json'];
export type DiscoveryDescriptionResponse = components['responses']['DiscoveryDescriptionResponse']['content']['application/json'];
export type DescribeAttributesResponse = components['responses']['DescribeAttributesResponse']['content']['application/json'];
export type TargetDiscoveryResponse = components['responses']['TargetDiscoveryResponse']['content']['application/json'];
export type DescribeTargetResponse = components['responses']['DescribeTargetResponse']['content']['application/json'];
