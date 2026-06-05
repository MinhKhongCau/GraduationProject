import { describe, it, expect } from 'vitest';

// Example utility functions to test
const formatCurrency = (amount: number): string => {
  return `$${amount.toFixed(2)}`;
};

const validateEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

const truncateText = (text: string, length: number): string => {
  if (text.length <= length) return text;
  return text.substring(0, length) + '...';
};

const calculateAge = (birthDate: Date): number => {
  const today = new Date();
  let age = today.getFullYear() - birthDate.getFullYear();
  const monthDiff = today.getMonth() - birthDate.getMonth();
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
    age--;
  }
  return age;
};

const capitalizeFirstLetter = (str: string): string => {
  if (str.length === 0) return str;
  return str.charAt(0).toUpperCase() + str.slice(1);
};

// Test Cases
describe('Utility Functions', () => {
  describe('formatCurrency', () => {
    it('should format a positive number as currency', () => {
      expect(formatCurrency(100)).toBe('$100.00');
    });

    it('should format a decimal number as currency', () => {
      expect(formatCurrency(99.5)).toBe('$99.50');
    });

    it('should format zero as currency', () => {
      expect(formatCurrency(0)).toBe('$0.00');
    });

    it('should format negative numbers as currency', () => {
      expect(formatCurrency(-50.25)).toBe('$-50.25');
    });
  });

  describe('validateEmail', () => {
    it('should validate a correct email', () => {
      expect(validateEmail('test@example.com')).toBe(true);
    });

    it('should invalidate an email without @', () => {
      expect(validateEmail('testexample.com')).toBe(false);
    });

    it('should invalidate an email without domain', () => {
      expect(validateEmail('test@')).toBe(false);
    });

    it('should invalidate an email with spaces', () => {
      expect(validateEmail('test @example.com')).toBe(false);
    });

    it('should validate emails with multiple subdomains', () => {
      expect(validateEmail('user@mail.example.co.uk')).toBe(true);
    });
  });

  describe('truncateText', () => {
    it('should not truncate text shorter than limit', () => {
      expect(truncateText('hello', 10)).toBe('hello');
    });

    it('should truncate text longer than limit', () => {
      expect(truncateText('hello world', 5)).toBe('hello...');
    });

    it('should truncate exactly at limit', () => {
      expect(truncateText('hello world', 5)).toBe('hello...');
    });

    it('should handle empty string', () => {
      expect(truncateText('', 5)).toBe('');
    });
  });

  describe('calculateAge', () => {
    it('should calculate age correctly', () => {
      const birthDate = new Date(2000, 0, 1); // Jan 1, 2000
      const age = calculateAge(birthDate);
      // Age should be around 26 (as of June 2026)
      expect(age).toBeGreaterThanOrEqual(25);
      expect(age).toBeLessThanOrEqual(26);
    });

    it('should calculate age for recent birth', () => {
      const birthDate = new Date(2025, 0, 1); // Jan 1, 2025
      const age = calculateAge(birthDate);
      expect(age).toBe(1);
    });

    it('should handle birthday today', () => {
      const today = new Date();
      const birthDate = new Date(today.getFullYear() - 30, today.getMonth(), today.getDate());
      expect(calculateAge(birthDate)).toBe(30);
    });
  });

  describe('capitalizeFirstLetter', () => {
    it('should capitalize first letter of lowercase string', () => {
      expect(capitalizeFirstLetter('hello')).toBe('Hello');
    });

    it('should handle already capitalized string', () => {
      expect(capitalizeFirstLetter('Hello')).toBe('Hello');
    });

    it('should handle single character', () => {
      expect(capitalizeFirstLetter('a')).toBe('A');
    });

    it('should handle empty string', () => {
      expect(capitalizeFirstLetter('')).toBe('');
    });

    it('should capitalize string with numbers', () => {
      expect(capitalizeFirstLetter('123abc')).toBe('123abc');
    });
  });
});
