'use strict';

/**
 * NexCore Official Node.js SDK — 公开入口.
 *
 * 一次 `require('@nexcore/sdk')` 即可拿到全部业务能力.
 *
 * @example
 *   const { Client, NexCoreError } = require('@nexcore/sdk');
 *   const client = new Client({ baseUrl: 'https://your-domain.com', ... });
 */

const { Client } = require('./src/client');
const { NexCoreError } = require('./src/errors');

module.exports = { Client, NexCoreError };
