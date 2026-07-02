const nextJest = require('next/jest');

const createJestConfig = nextJest({ dir: './' });

module.exports = createJestConfig({
  testEnvironment: 'jsdom',
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@aws-sdk/client-s3$': '<rootDir>/src/__tests__/__mocks__/aws-sdk-s3.ts',
  },
  testMatch: ['<rootDir>/src/__tests__/**/*.test.{ts,tsx}'],
});
