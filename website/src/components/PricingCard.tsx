import { motion } from 'framer-motion';
import { Check, Star } from 'lucide-react';
import Button from './Button';

interface PricingCardProps {
  title: string;
  price: string;
  period: string;
  description: string;
  features: string[];
  buttonText: string;
  popular: boolean;
}

const PricingCard = ({ title, price, period, description, features, buttonText, popular }: PricingCardProps) => {
  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      whileHover={{ y: -5 }}
      className={`relative bg-white rounded-2xl shadow-lg p-8 hover:shadow-xl transition-all duration-300 ${
        popular ? 'ring-2 ring-blue-600 scale-105' : 'border border-gray-200'
      }`}
    >
      {popular && (
        <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
          <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white px-4 py-1 rounded-full text-sm font-semibold flex items-center gap-1">
            <Star className="w-4 h-4" />
            Most Popular
          </div>
        </div>
      )}
      
      <div className="text-center mb-8">
        <h3 className="text-2xl font-bold text-gray-900 mb-2">{title}</h3>
        <div className="mb-4">
          <span className="text-4xl font-bold text-gray-900">{price}</span>
          <span className="text-gray-600 ml-1">{period}</span>
        </div>
        <p className="text-gray-600">{description}</p>
      </div>
      
      <div className="mb-8">
        <ul className="space-y-3">
          {features.map((feature, index) => (
            <li key={index} className="flex items-start">
              <Check className="w-5 h-5 text-green-500 mr-3 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700">{feature}</span>
            </li>
          ))}
        </ul>
      </div>
      
      <Button 
        className={`w-full ${popular ? 'bg-blue-600 hover:bg-blue-700' : ''}`}
        variant={popular ? 'default' : 'outline'}
      >
        {buttonText}
      </Button>
    </motion.div>
  );
};

export default PricingCard;