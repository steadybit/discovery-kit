// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

import { router as indexRouter } from './routes/index';
import { router as discoveryRouter } from './routes/discovery';
import { router as targettypesRouter } from './routes/targettypes';
import { router as targetattributesRouter } from './routes/targetattributes';
import express from 'express';
import cors from 'cors';

const app = express();
const port = 8085;

app.use(express.json());

app.use(cors());
app.use(indexRouter);
app.use(discoveryRouter);
app.use(targettypesRouter);
app.use(targetattributesRouter);

app.listen(port, () => {
	console.log(`Discovery implementation listening on ${port}.`);
	console.log();
	console.log(`Discovery extension can be accessed via GET http://127.0.0.1:${port}/`);
});
