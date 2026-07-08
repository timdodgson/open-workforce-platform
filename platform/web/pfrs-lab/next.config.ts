import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone',
  async redirects() {
    return [
      { source: '/predictions', destination: '/intelligence?tab=predictions', permanent: false },
      { source: '/intelligence/policies', destination: '/intelligence?tab=policies', permanent: false },
      { source: '/intelligence/validation', destination: '/intelligence?tab=si-validation', permanent: false },
      { source: '/intelligence/learning', destination: '/intelligence?tab=learning', permanent: false },
      { source: '/intelligence/promotion', destination: '/intelligence?tab=policies', permanent: false },
      { source: '/intelligence/counterfactual', destination: '/intelligence?tab=decisions', permanent: false },
      { source: '/sa', destination: '/runs', permanent: false },
      { source: '/summary', destination: '/runs', permanent: false },
      { source: '/workers', destination: '/runs', permanent: false },
      { source: '/tree', destination: '/runs', permanent: false },
      { source: '/search', destination: '/runs', permanent: false },
      { source: '/diversity', destination: '/runs', permanent: false },
    ];
  },
};

export default nextConfig;
