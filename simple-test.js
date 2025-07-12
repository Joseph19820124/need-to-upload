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
        'User-Agent': 'Simple-Test-Client',
        ...options.headers
      }
    }, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          resolve({
            statusCode: res.statusCode,
            body: data,
            json: JSON.parse(data)
          });
        } catch (e) {
          resolve({ statusCode: res.statusCode, body: data, json: null });
        }
      });
    });
    
    req.on('error', reject);
    if (options.body) req.write(options.body);
    req.end();
  });
}

async function runTests() {
  console.log('🧪 Simple MCP Server Test');
  console.log(`📡 Server: ${SERVER_URL}`);
  console.log('=' .repeat(50));

  try {
    // Test 1: Health Check
    console.log('⏳ Testing health endpoint...');
    const health = await request('/api/v1/health');
    console.log(`✅ Health: ${health.statusCode === 200 ? 'PASS' : 'FAIL'}`);
    if (health.json) console.log(`   ${JSON.stringify(health.json)}`);

    // Test 2: Connect
    console.log('⏳ Testing connect endpoint...');
    const connect = await request('/api/v1/connect', {
      method: 'POST',
      body: JSON.stringify({
        clientInfo: { name: 'simple-test', version: '1.0.0' }
      })
    });
    
    console.log(`✅ Connect: ${connect.statusCode === 200 ? 'PASS' : 'FAIL'}`);
    if (connect.json) {
      console.log(`   Session: ${connect.json.sessionId}`);
      console.log(`   Server: ${connect.json.serverInfo?.name}`);
      
      // Test 3: RPC with session
      if (connect.json.sessionId) {
        console.log('⏳ Testing RPC with session...');
        const rpc = await request('/api/v1/rpc', {
          method: 'POST',
          headers: { 'X-Session-ID': connect.json.sessionId },
          body: JSON.stringify({
            jsonrpc: '2.0',
            id: 1,
            method: 'tools/list',
            params: {}
          })
        });
        
        console.log(`✅ RPC: ${rpc.statusCode === 200 ? 'PASS' : 'FAIL'}`);
        if (rpc.json && rpc.json.result) {
          console.log(`   Tools: ${rpc.json.result.tools?.length || 0} available`);
        }
      }
    }

  } catch (error) {
    console.log(`❌ Error: ${error.message}`);
  }
  
  console.log('=' .repeat(50));
  console.log('🎉 Test completed!');
}

runTests();