import { Footer, Layout, Navbar } from 'nextra-theme-docs'
import { Banner, Head } from 'nextra/components'
import { getPageMap } from 'nextra/page-map'
import 'nextra-theme-docs/style.css'
import '../../styles.css';
import { GetStartedButton } from '../components/GetStartedModal';
 
export const metadata = {
  // Define your metadata here
  // For more information on metadata API, see: https://nextjs.org/docs/app/building-your-application/optimizing/metadata
  description: ' InfraSpec is a tool for testing your cloud infrastructure in plain English, no code required.',
  metadataBase: new URL('https://infraspec.sh'),
  keywords: [
    'InfraSpec',
    'Terratest',
    'IaC',
    'Terraform',
    'Infrastructure Testing Tool',
    'Testing',
    'Testing Tool',
    'Testing Framework',
    'Testing Infrastructure',
    'Testing Terraform',
    'Testing Infrastructure as Code',
    'Testing Infrastructure as Code with Terraform',
    'Testing Infrastructure as Code with Terratest',
    'Testing Infrastructure as Code with InfraSpec',
  ],
  title: {
    default: 'InfraSpec - Test Cloud Infrastructure',
    template: '%s | InfraSpec',
  },
  openGraph: {
    // https://github.com/vercel/next.js/discussions/50189#discussioncomment-10826632
    url: './',
    siteName: 'InfraSpec',
    locale: 'en_US',
    type: 'website'
  },
  other: {
    'msapplication-TileColor': '#fff'
  },
  twitter: {
    site: 'https://infraspec.sh',
    creator: '@_rjm_'
  },
  alternates: {
    // https://github.com/vercel/next.js/discussions/50189#discussioncomment-10826632
    canonical: './'
  }
}
 
const banner = <Banner storageKey="some-key">Nextra 4.0 is released 🎉</Banner>
const navbar = (
  <Navbar
    logo={
      <>
        <img src="/infraspec_logo_512.png" width="50px" loading="lazy" />
        <span className="mx-2 font-extrabold hidden md:inline select-none">
          InfraSpec
        </span>
        <span className="text-gray-600 font-normal hidden lg:!inline whitespace-no-wrap">
          Test Cloud Infrastructure
        </span>
      </>
    }
    projectLink="https://github.com/robmorgan/infraspec"
  >
    <a href="/early-access" className="hidden sm:inline-block text-sm hover:text-purple-600 transition-colors">
      Early Access
    </a>
    <GetStartedButton className="hidden sm:inline-block bg-purple-600 hover:bg-purple-700 text-white font-semibold py-1.5 px-4 rounded-lg text-sm transition-colors">
      Get Started
    </GetStartedButton>
  </Navbar>
)
const footer = <Footer>{new Date().getFullYear()} © Rob Morgan.</Footer>
 
export default async function RootLayout({ children }) {
  return (
    <html
      // Not required, but good for SEO
      lang="en"
      // Required to be set
      dir="ltr"
      // Suggested by `next-themes` package https://github.com/pacocoursey/next-themes#with-app
      suppressHydrationWarning
    >
      <Head
      // ... Your additional head options
      >
        {/* Your additional tags should be passed as `children` of `<Head>` element */}
      </Head>
      <body>
        <Layout
          navbar={navbar}
          pageMap={await getPageMap()}
          docsRepositoryBase="https://github.com/robmorgan/infraspec/tree/main/website"
          footer={footer}
          // ... Your additional layout options
        >
          {children}
        </Layout>
      </body>
    </html>
  )
}