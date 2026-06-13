#!/usr/bin/env node

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function countKeys(obj, prefix = '') {
  let count = 0;
  let keys = [];
  
  for (const [key, value] of Object.entries(obj)) {
    const fullKey = prefix ? `${prefix}.${key}` : key;
    
    if (typeof value === 'object' && value !== null) {
      const result = countKeys(value, fullKey);
      count += result.count;
      keys.push(...result.keys);
    } else {
      count++;
      keys.push(fullKey);
    }
  }
  
  return { count, keys };
}

function checkI18nConsistency() {
  const enPath = path.join(__dirname, '../src/locales/en.json');
  const zhPath = path.join(__dirname, '../src/locales/zh.json');
  
  try {
    const enContent = JSON.parse(fs.readFileSync(enPath, 'utf8'));
    const zhContent = JSON.parse(fs.readFileSync(zhPath, 'utf8'));
    
    const enResult = countKeys(enContent);
    const zhResult = countKeys(zhContent);
    
    console.log('🌐 Internationalization Consistency Check');
    console.log('==========================================');
    console.log(`📄 English (en.json): ${enResult.count} keys`);
    console.log(`📄 Chinese (zh.json): ${zhResult.count} keys`);
    
    if (enResult.count !== zhResult.count) {
      console.error('❌ ERROR: Translation files have different key counts!');
      
      const enKeys = new Set(enResult.keys);
      const zhKeys = new Set(zhResult.keys);
      
      const missingInZh = [...enKeys].filter(key => !zhKeys.has(key));
      const missingInEn = [...zhKeys].filter(key => !enKeys.has(key));
      
      if (missingInZh.length > 0) {
        console.error('🔴 Keys missing in zh.json:');
        missingInZh.forEach(key => console.error(`   - ${key}`));
      }
      
      if (missingInEn.length > 0) {
        console.error('🔴 Keys missing in en.json:');
        missingInEn.forEach(key => console.error(`   - ${key}`));
      }
      
      process.exit(1);
    } else {
      console.log('✅ SUCCESS: Both translation files have consistent key counts');
      console.log('🎉 All translation keys are properly synchronized');
    }
  } catch (error) {
    console.error('❌ ERROR: Failed to read translation files');
    console.error(error.message);
    process.exit(1);
  }
}

checkI18nConsistency();