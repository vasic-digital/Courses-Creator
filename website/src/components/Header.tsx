import { useState } from 'react';
import Link from 'next/link';
import { motion } from 'framer-motion';
import { Menu, X, ChevronDown, Play, BookOpen, Users, Zap } from 'lucide-react';
import Button from './Button';

const Header = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isProductsOpen, setIsProductsOpen] = useState(false);
  const [isResourcesOpen, setIsResourcesOpen] = useState(false);

  const navigation = [
    {
      name: 'Products',
      href: '#',
      icon: <Zap className="w-4 h-4" />,
      children: [
        { name: 'Course Creator', href: '/products/course-creator', description: 'Create professional video courses' },
        { name: 'Player Apps', href: '/products/player-apps', description: 'Multi-platform learning experience' },
        { name: 'Analytics', href: '/products/analytics', description: 'Detailed learning analytics' },
        { name: 'API', href: '/products/api', description: 'Integrate with your systems' }
      ]
    },
    {
      name: 'Resources',
      href: '#',
      icon: <BookOpen className="w-4 h-4" />,
      children: [
        { name: 'Documentation', href: '/docs', description: 'Comprehensive guides' },
        { name: 'Tutorials', href: '/tutorials', description: 'Step-by-step tutorials' },
        { name: 'Blog', href: '/blog', description: 'Latest insights and tips' },
        { name: 'Community', href: '/community', description: 'Connect with other creators' }
      ]
    },
    { name: 'Pricing', href: '/pricing' },
    { name: 'Enterprise', href: '/enterprise' },
    { name: 'About', href: '/about' }
  ];

  return (
    <header className="fixed top-0 w-full bg-white/95 backdrop-blur-sm border-b border-gray-200 z-50">
      <nav className="container mx-auto px-4">
        <div className="flex justify-between items-center h-16">
          {/* Logo */}
          <Link href="/" className="flex items-center space-x-2">
            <div className="w-8 h-8 bg-gradient-to-br from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
              <Play className="w-5 h-5 text-white" />
            </div>
            <span className="text-xl font-bold text-gray-900">Course Creator</span>
          </Link>

          {/* Desktop Navigation */}
          <div className="hidden lg:flex items-center space-x-8">
            {navigation.map((item) => (
              <div key={item.name} className="relative">
                {item.children ? (
                  <div
                    className="flex items-center space-x-1 text-gray-700 hover:text-blue-600 cursor-pointer py-2"
                    onMouseEnter={() => item.name === 'Products' ? setIsProductsOpen(true) : setIsResourcesOpen(true)}
                    onMouseLeave={() => item.name === 'Products' ? setIsProductsOpen(false) : setIsResourcesOpen(false)}
                  >
                    {item.icon}
                    <span className="font-medium">{item.name}</span>
                    <ChevronDown className="w-4 h-4" />
                  </div>
                ) : (
                  <Link
                    href={item.href}
                    className="text-gray-700 hover:text-blue-600 font-medium py-2"
                  >
                    {item.name}
                  </Link>
                )}

                {/* Dropdown Menu */}
                {item.children && (
                  <motion.div
                    initial={{ opacity: 0, y: -10 }}
                    animate={{ 
                      opacity: item.name === 'Products' ? (isProductsOpen ? 1 : 0) : (isResourcesOpen ? 1 : 0),
                      y: item.name === 'Products' ? (isProductsOpen ? 0 : -10) : (isResourcesOpen ? 0 : -10)
                    }}
                    className="absolute top-full left-0 mt-2 w-80 bg-white rounded-lg shadow-lg border border-gray-200 py-2"
                    onMouseEnter={() => item.name === 'Products' ? setIsProductsOpen(true) : setIsResourcesOpen(true)}
                    onMouseLeave={() => item.name === 'Products' ? setIsProductsOpen(false) : setIsResourcesOpen(false)}
                  >
                    {item.children.map((child) => (
                      <Link
                        key={child.name}
                        href={child.href}
                        className="block px-4 py-3 hover:bg-gray-50"
                      >
                        <div className="font-medium text-gray-900">{child.name}</div>
                        <div className="text-sm text-gray-600">{child.description}</div>
                      </Link>
                    ))}
                  </motion.div>
                )}
              </div>
            ))}
          </div>

          {/* Desktop CTA */}
          <div className="hidden lg:flex items-center space-x-4">
            <Link href="/signin">
              <Button variant="ghost">Sign In</Button>
            </Link>
            <Link href="/signup">
              <Button>Start Free Trial</Button>
            </Link>
          </div>

          {/* Mobile Menu Button */}
          <button
            className="lg:hidden p-2 rounded-md text-gray-700 hover:text-blue-600"
            onClick={() => setIsMenuOpen(!isMenuOpen)}
          >
            {isMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>

        {/* Mobile Navigation */}
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: isMenuOpen ? 1 : 0, height: isMenuOpen ? 'auto' : 0 }}
          className="lg:hidden overflow-hidden"
        >
          <div className="py-4 space-y-2">
            {navigation.map((item) => (
              <div key={item.name}>
                {item.children ? (
                  <div>
                    <button
                      className="flex items-center justify-between w-full px-3 py-2 text-gray-700 hover:text-blue-600"
                      onClick={() => item.name === 'Products' ? setIsProductsOpen(!isProductsOpen) : setIsResourcesOpen(!isResourcesOpen)}
                    >
                      <div className="flex items-center space-x-2">
                        {item.icon}
                        <span className="font-medium">{item.name}</span>
                      </div>
                      <ChevronDown className={`w-4 h-4 transition-transform ${item.name === 'Products' ? (isProductsOpen ? 'rotate-180' : '') : (isResourcesOpen ? 'rotate-180' : '')}`} />
                    </button>
                    <motion.div
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ 
                        opacity: item.name === 'Products' ? (isProductsOpen ? 1 : 0) : (isResourcesOpen ? 1 : 0),
                        height: item.name === 'Products' ? (isProductsOpen ? 'auto' : 0) : (isResourcesOpen ? 'auto' : 0)
                      }}
                      className="overflow-hidden"
                    >
                      <div className="pl-6 pr-3 py-2 space-y-1">
                        {item.children.map((child) => (
                          <Link
                            key={child.name}
                            href={child.href}
                            className="block px-3 py-2 text-gray-600 hover:text-blue-600"
                            onClick={() => setIsMenuOpen(false)}
                          >
                            <div className="font-medium">{child.name}</div>
                            <div className="text-sm text-gray-500">{child.description}</div>
                          </Link>
                        ))}
                      </div>
                    </motion.div>
                  </div>
                ) : (
                  <Link
                    href={item.href}
                    className="block px-3 py-2 text-gray-700 hover:text-blue-600 font-medium"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    {item.name}
                  </Link>
                )}
              </div>
            ))}
            
            <div className="pt-4 mt-4 border-t border-gray-200 space-y-2">
              <Link href="/signin" onClick={() => setIsMenuOpen(false)}>
                <Button variant="ghost" className="w-full justify-start">Sign In</Button>
              </Link>
              <Link href="/signup" onClick={() => setIsMenuOpen(false)}>
                <Button className="w-full">Start Free Trial</Button>
              </Link>
            </div>
          </div>
        </motion.div>
      </nav>
    </header>
  );
};

export default Header;