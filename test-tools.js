#!/usr/bin/env node

const https = require('https');

const SERVER_URL = 'https://calm-benevolence-production.up.railway.app';

function request(path, options = {}) {
  return new Promise((resolve, reject) => {
    const fullUrl = SERVER_URL + path;
    const req = https.request(fullUrl, {
      method: options.method || 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      }
    }, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          resolve({ statusCode: res.statusCode, json: JSON.parse(data) });
        } catch (e) {
          resolve({ statusCode: res.statusCode, body: data });
        }
      });
    });
    
    req.on('error', reject);
    if (options.body) req.write(options.body);
    req.end();
  });
}

async function testTools() {
  try {
    // Connect first
    const connect = await request('/api/v1/connect', {
      method: 'POST',
      body: JSON.stringify({
        clientInfo: { name: 'tool-test', version: '1.0.0' }
      })
    });

    if (connect.json && connect.json.sessionId) {
      console.log('✅ Connected with session:', connect.json.sessionId);
      
      // List tools
      const tools = await request('/api/v1/rpc', {
        method: 'POST',
        headers: { 'X-Session-ID': connect.json.sessionId },
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: 1,
          method: 'tools/list',
          params: {}
        })
      });
      
      console.log('🔧 Available Tools:');
      console.log(JSON.stringify(tools.json, null, 2));
      
      // List resources
      const resources = await request('/api/v1/rpc', {
        method: 'POST',
        headers: { 'X-Session-ID': connect.json.sessionId },
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: 2,
          method: 'resources/list',
          params: {}
        })
      });
      
      console.log('📚 Available Resources:');
      console.log(JSON.stringify(resources.json, null, 2));
    }
  } catch (error) {
    console.error('Error:', error.message);
  }
}

testTools();