import { GetStaticProps } from 'next';
import Head from 'next/head';
import { motion } from 'framer-motion';
import { Play, BookOpen, Users, Zap, Shield, Globe, ArrowRight, CheckCircle, Star } from 'lucide-react';
import Header from '../components/Header';
import Footer from '../components/Footer';
import Button from '../components/Button';
import TestimonialCard from '../components/TestimonialCard';
import FeatureCard from '../components/FeatureCard';
import PricingCard from '../components/PricingCard';

interface HomePageProps {
  testimonials: Array<{
    id: string;
    name: string;
    role: string;
    company: string;
    content: string;
    rating: number;
    avatar: string;
  }>;
}

export default function HomePage({ testimonials }: HomePageProps) {
  return (
    <>
      <Head>
        <title>Course Creator - Create Professional Video Courses with AI</title>
        <meta name="description" content="Create stunning video courses with AI-powered tools. Professional quality, multi-platform support, and comprehensive analytics." />
        <meta name="keywords" content="course creator, video courses, e-learning, AI education, online learning" />
        <meta property="og:title" content="Course Creator - Professional Video Course Platform" />
        <meta property="og:description" content="Create professional video courses with AI-powered tools and multi-platform support." />
        <meta property="og:type" content="website" />
        <meta property="og:url" content="https://coursecreator.com" />
        <meta name="twitter:card" content="summary_large_image" />
        <meta name="twitter:title" content="Course Creator - Professional Video Course Platform" />
        <meta name="twitter:description" content="Create professional video courses with AI-powered tools." />
      </Head>

      <div className="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
        <Header />
        
        {/* Hero Section */}
        <section className="relative overflow-hidden py-20 px-4">
          <div className="container mx-auto max-w-6xl">
            <div className="grid lg:grid-cols-2 gap-12 items-center">
              <motion.div
                initial={{ opacity: 0, x: -50 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.6 }}
              >
                <h1 className="text-5xl lg:text-6xl font-bold text-gray-900 leading-tight mb-6">
                  Create Professional
                  <span className="text-blue-600"> Video Courses</span>
                  <br />with AI Power
                </h1>
                <p className="text-xl text-gray-600 mb-8 leading-relaxed">
                  Transform your expertise into engaging video courses with our AI-powered platform. 
                  Professional quality, multi-platform support, and comprehensive analytics - all in one place.
                </p>
                <div className="flex flex-col sm:flex-row gap-4 mb-8">
                  <Button size="lg" className="bg-blue-600 hover:bg-blue-700">
                    <Play className="w-5 h-5 mr-2" />
                    Start Free Trial
                  </Button>
                  <Button variant="outline" size="lg">
                    <BookOpen className="w-5 h-5 mr-2" />
                    Watch Demo
                  </Button>
                </div>
                <div className="flex items-center gap-6 text-sm text-gray-600">
                  <div className="flex items-center gap-1">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    <span>14-day free trial</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    <span>No credit card required</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <CheckCircle className="w-4 h-4 text-green-500" />
                    <span>Cancel anytime</span>
                  </div>
                </div>
              </motion.div>
              
              <motion.div
                initial={{ opacity: 0, x: 50 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.6, delay: 0.2 }}
                className="relative"
              >
                <div className="relative bg-white rounded-2xl shadow-2xl p-8">
                  <div className="aspect-video bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                    <Play className="w-16 h-16 text-white" />
                  </div>
                  <div className="mt-6 space-y-4">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-gray-600">Course Progress</span>
                      <span className="text-sm font-bold text-blue-600">85%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-blue-600 h-2 rounded-full" style={{ width: '85%' }}></div>
                    </div>
                    <div className="grid grid-cols-3 gap-4 text-center">
                      <div>
                        <div className="text-2xl font-bold text-gray-900">2.5K</div>
                        <div className="text-xs text-gray-600">Students</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-gray-900">4.8</div>
                        <div className="text-xs text-gray-600">Rating</div>
                      </div>
                      <div>
                        <div className="text-2xl font-bold text-gray-900">98%</div>
                        <div className="text-xs text-gray-600">Completion</div>
                      </div>
                    </div>
                  </div>
                </div>
                
                {/* Floating elements */}
                <div className="absolute -top-4 -right-4 bg-yellow-400 text-yellow-900 px-4 py-2 rounded-full text-sm font-bold shadow-lg">
                  AI Powered
                </div>
                <div className="absolute -bottom-4 -left-4 bg-green-500 text-white px-4 py-2 rounded-full text-sm font-bold shadow-lg">
                  1080p Quality
                </div>
              </motion.div>
            </div>
          </div>
        </section>

        {/* Features Section */}
        <section className="py-20 px-4 bg-white">
          <div className="container mx-auto max-w-6xl">
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6 }}
              className="text-center mb-16"
            >
              <h2 className="text-4xl font-bold text-gray-900 mb-4">
                Everything You Need to Create Amazing Courses
              </h2>
              <p className="text-xl text-gray-600 max-w-3xl mx-auto">
                Professional tools powered by AI to help you create, manage, and deliver engaging video courses across all platforms.
              </p>
            </motion.div>

            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
              <FeatureCard
                icon={<Zap className="w-8 h-8" />}
                title="AI-Powered Creation"
                description="Generate course content, voiceovers, and subtitles automatically with advanced AI technology."
              />
              <FeatureCard
                icon={<Globe className="w-8 h-8" />}
                title="Multi-Platform Support"
                description="Publish to web, mobile apps, and LMS platforms with a single click."
              />
              <FeatureCard
                icon={<Users className="w-8 h-8" />}
                title="Collaboration Tools"
                description="Work with your team in real-time with role-based permissions and version control."
              />
              <FeatureCard
                icon={<Shield className="w-8 h-8" />}
                title="Enterprise Security"
                description="Bank-level encryption, GDPR compliance, and comprehensive audit trails."
              />
              <FeatureCard
                icon={<BookOpen className="w-8 h-8" />}
                title="Rich Analytics"
                description="Track engagement, completion rates, and learning outcomes with detailed analytics."
              />
              <FeatureCard
                icon={<Play className="w-8 h-8" />}
                title="Professional Quality"
                description="1080p video, crystal-clear audio, and professional editing tools built-in."
              />
            </div>
          </div>
        </section>

        {/* Testimonials Section */}
        <section className="py-20 px-4 bg-gray-50">
          <div className="container mx-auto max-w-6xl">
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6 }}
              className="text-center mb-16"
            >
              <h2 className="text-4xl font-bold text-gray-900 mb-4">
                Loved by Course Creators Worldwide
              </h2>
              <p className="text-xl text-gray-600 max-w-3xl mx-auto">
                Join thousands of educators and businesses creating amazing learning experiences.
              </p>
            </motion.div>

            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
              {testimonials.map((testimonial, index) => (
                <TestimonialCard key={testimonial.id} testimonial={testimonial} index={index} />
              ))}
            </div>
          </div>
        </section>

        {/* Pricing Section */}
        <section className="py-20 px-4 bg-white">
          <div className="container mx-auto max-w-6xl">
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6 }}
              className="text-center mb-16"
            >
              <h2 className="text-4xl font-bold text-gray-900 mb-4">
                Simple, Transparent Pricing
              </h2>
              <p className="text-xl text-gray-600 max-w-3xl mx-auto">
                Choose the perfect plan for your needs. All plans include our core features.
              </p>
            </motion.div>

            <div className="grid md:grid-cols-3 gap-8">
              <PricingCard
                title="Starter"
                price="$29"
                period="/month"
                description="Perfect for individual creators getting started"
                features={[
                  'Up to 3 courses',
                  '100 students per course',
                  'Basic analytics',
                  'Email support',
                  'Standard video quality (720p)',
                  'Mobile app access'
                ]}
                buttonText="Start Free Trial"
                popular={false}
              />
              <PricingCard
                title="Professional"
                price="$99"
                period="/month"
                description="Ideal for growing educational businesses"
                features={[
                  'Unlimited courses',
                  '1,000 students per course',
                  'Advanced analytics',
                  'Priority support',
                  'HD video quality (1080p)',
                  'Collaboration tools',
                  'API access',
                  'Custom branding'
                ]}
                buttonText="Start Free Trial"
                popular={true}
              />
              <PricingCard
                title="Enterprise"
                price="Custom"
                period=""
                description="For large organizations with specific needs"
                features={[
                  'Everything in Professional',
                  'Unlimited students',
                  'Dedicated account manager',
                  'SLA guarantee',
                  'Custom integrations',
                  'White-label solution',
                  'On-premise deployment',
                  'Advanced security features'
                ]}
                buttonText="Contact Sales"
                popular={false}
              />
            </div>
          </div>
        </section>

        {/* CTA Section */}
        <section className="py-20 px-4 bg-gradient-to-r from-blue-600 to-purple-600">
          <div className="container mx-auto max-w-4xl text-center">
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6 }}
            >
              <h2 className="text-4xl font-bold text-white mb-6">
                Ready to Create Your First Course?
              </h2>
              <p className="text-xl text-blue-100 mb-8">
                Join thousands of course creators who are already using Course Creator to deliver amazing learning experiences.
              </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <Button size="lg" className="bg-white text-blue-600 hover:bg-gray-100">
                  Start Your Free Trial
                  <ArrowRight className="w-5 h-5 ml-2" />
                </Button>
                <Button variant="outline" size="lg" className="border-white text-white hover:bg-white hover:text-blue-600">
                  Schedule a Demo
                </Button>
              </div>
              <div className="mt-8 flex items-center justify-center gap-8 text-blue-100">
                <div className="flex items-center gap-1">
                  <CheckCircle className="w-5 h-5" />
                  <span>No setup fees</span>
                </div>
                <div className="flex items-center gap-1">
                  <CheckCircle className="w-5 h-5" />
                  <span>14-day trial</span>
                </div>
                <div className="flex items-center gap-1">
                  <CheckCircle className="w-5 h-5" />
                  <span>Cancel anytime</span>
                </div>
              </div>
            </motion.div>
          </div>
        </section>

        <Footer />
      </div>
    </>
  );
}

export const getStaticProps: GetStaticProps = async () => {
  const testimonials = [
    {
      id: '1',
      name: 'Sarah Johnson',
      role: 'Online Educator',
      company: 'Tech Academy',
      content: 'Course Creator has transformed how I create and deliver content. The AI-powered features save me hours every week, and my students love the professional quality.',
      rating: 5,
      avatar: '/images/testimonials/sarah.jpg'
    },
    {
      id: '2',
      name: 'Michael Chen',
      role: 'Training Director',
      company: 'Global Corp',
      content: 'We\'ve tried many course platforms, but Course Creator stands out with its enterprise features and excellent support. Our training programs have never been more effective.',
      rating: 5,
      avatar: '/images/testimonials/michael.jpg'
    },
    {
      id: '3',
      name: 'Emily Rodriguez',
      role: 'Course Creator',
      company: 'Creative Learning',
      content: 'The collaboration tools are game-changing. Our team can work together seamlessly, and the analytics help us continuously improve our courses.',
      rating: 5,
      avatar: '/images/testimonials/emily.jpg'
    }
  ];

  return {
    props: {
      testimonials
    },
    revalidate: 60 // Revalidate every minute
  };
};