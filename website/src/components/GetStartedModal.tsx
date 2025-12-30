'use client';

import { useState, useEffect, useCallback } from 'react';

interface GetStartedModalProps {
  isOpen: boolean;
  onClose: () => void;
}

function GetStartedModal({ isOpen, onClose }: GetStartedModalProps) {
  const [copied, setCopied] = useState(false);
  const curlCommand = 'curl -fsSL https://infraspec.sh/install.sh | bash';

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(curlCommand);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = curlCommand;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }, [curlCommand]);

  // Handle escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isOpen, onClose]);

  // Prevent body scroll when modal is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      onClick={onClose}
    >
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" />

      {/* Modal */}
      <div
        className="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-xl max-h-[90vh] overflow-auto"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Close button */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors z-10"
          aria-label="Close modal"
        >
          <svg
            className="w-6 h-6"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>

        <div>
          {/* CLI Installation */}
          <div className="p-8">
            <div className="flex items-center gap-3 mb-4">
              <div className="p-2 bg-gray-100 dark:bg-gray-800 rounded-lg">
                <svg
                  className="w-6 h-6 text-gray-700 dark:text-gray-300"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                  />
                </svg>
              </div>
              <h3 className="text-xl font-bold text-gray-900 dark:text-white">
                Install via Terminal
              </h3>
            </div>

            <p className="text-gray-600 dark:text-gray-400 mb-6">
              Get started in seconds with a single command. Works on macOS, Linux, and WSL.
            </p>

            {/* Code block */}
            <div className="relative group">
              <div className="bg-gray-900 dark:bg-gray-950 rounded-lg p-4 font-mono text-sm overflow-x-auto">
                <code className="text-green-400">{curlCommand}</code>
              </div>
              <button
                onClick={handleCopy}
                className="absolute top-2 right-2 p-2 bg-gray-700 hover:bg-gray-600 rounded-md transition-colors opacity-0 group-hover:opacity-100 focus:opacity-100"
                aria-label="Copy to clipboard"
              >
                {copied ? (
                  <svg
                    className="w-5 h-5 text-green-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                ) : (
                  <svg
                    className="w-5 h-5 text-gray-300"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                    />
                  </svg>
                )}
              </button>
              {copied && (
                <span className="absolute top-2 right-12 text-xs text-green-400 bg-gray-800 px-2 py-1 rounded">
                  Copied!
                </span>
              )}
            </div>

            {/* Alternative methods */}
            <div className="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
              <p className="text-sm text-gray-500 dark:text-gray-400 mb-3">
                Or install using:
              </p>
              <div className="flex flex-wrap gap-3">
                <a
                  href="/docs/getting-started#homebrew"
                  className="inline-flex items-center gap-2 text-sm text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300"
                >
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M5.547 2.251A12.003 12.003 0 0 0 .457 8.87c-.25.818-.25 1.693-.25 3.444 0 1.75 0 2.625.25 3.444a12.003 12.003 0 0 0 5.09 6.617c.727.423 1.529.635 3.094.847 1.565.212 2.348.318 3.107.148a12.003 12.003 0 0 0 7.795-5.09c.423-.727.635-1.529.847-3.094.212-1.565.318-2.348.148-3.107a12.003 12.003 0 0 0-5.09-7.795c-.727-.423-1.529-.635-3.094-.847-1.565-.212-2.348-.318-3.107-.148a12.003 12.003 0 0 0-3.7 1.962Z" />
                  </svg>
                  Homebrew
                </a>
                <a
                  href="https://github.com/robmorgan/infraspec/releases"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-2 text-sm text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300"
                >
                  <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12Z" />
                  </svg>
                  GitHub Releases
                </a>
                <a
                  href="/docs/getting-started#building-from-source"
                  className="inline-flex items-center gap-2 text-sm text-purple-600 hover:text-purple-700 dark:text-purple-400 dark:hover:text-purple-300"
                >
                  <svg
                    className="w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
                    />
                  </svg>
                  Build from Source
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

interface GetStartedButtonProps {
  className?: string;
  children?: React.ReactNode;
}

export function GetStartedButton({ className = '', children }: GetStartedButtonProps) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className={className || `inline-block bg-purple-600 hover:bg-purple-700 text-white font-bold py-3 px-8 rounded-lg text-lg transition-colors`}
      >
        {children || 'Get Started'}
      </button>
      <GetStartedModal isOpen={isOpen} onClose={() => setIsOpen(false)} />
    </>
  );
}

export default GetStartedModal;
