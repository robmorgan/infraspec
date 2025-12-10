import DOMPurify from 'isomorphic-dompurify';

// RFC 5322 compliant email regex
const EMAIL_REGEX = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;

/**
 * Sanitize user input by stripping all HTML tags
 */
export function sanitize(text: string): string {
  return DOMPurify.sanitize(text, { ALLOWED_TAGS: [] });
}

/**
 * Validate email format with strict RFC 5322 regex
 * Also checks for header injection characters and max length
 */
export function isValidEmail(email: string): boolean {
  if (!email || email.length > 254) return false;
  // Prevent email header injection via newline, carriage return, or null bytes
  if (/[\r\n\x00]/.test(email)) return false;
  return EMAIL_REGEX.test(email);
}

// Input length limits
export const INPUT_LIMITS = {
  name: 100,
  email: 254,
  message: 5000,
  comments: 2000,
} as const;
