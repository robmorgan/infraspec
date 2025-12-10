import { Resend } from 'resend';
import { NextRequest, NextResponse } from 'next/server';
import DOMPurify from 'isomorphic-dompurify';

const resend = new Resend(process.env.RESEND_API_KEY);

// RFC 5322 compliant email regex
const EMAIL_REGEX = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;

function sanitize(text: string): string {
  return DOMPurify.sanitize(text, { ALLOWED_TAGS: [] });
}

function isValidEmail(email: string): boolean {
  if (!email || email.length > 254) return false;
  return EMAIL_REGEX.test(email);
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { name, email, comments } = body;

    // Validate required fields
    if (!name || !email) {
      return NextResponse.json(
        { error: 'Name and email are required' },
        { status: 400 }
      );
    }

    // Validate email format with strict regex
    if (!isValidEmail(email)) {
      return NextResponse.json(
        { error: 'Invalid email format' },
        { status: 400 }
      );
    }

    // Sanitize all user input
    const safeName = sanitize(name);
    const safeEmail = sanitize(email);
    const safeComments = comments ? sanitize(comments) : '';

    // Send email using Resend
    const data = await resend.emails.send({
      from: 'InfraSpec Early Access <noreply@infraspec.sh>',
      to: 'rob@brightfame.co',
      replyTo: safeEmail,
      subject: `Virtual Cloud Early Access Request from ${safeName}`,
      html: `
        <h2>New Virtual Cloud Early Access Request</h2>
        <p><strong>Name:</strong> ${safeName}</p>
        <p><strong>Email:</strong> ${safeEmail}</p>
        ${safeComments ? `<p><strong>Comments:</strong></p><p>${safeComments.replace(/\n/g, '<br>')}</p>` : ''}
      `,
    });

    return NextResponse.json(
      { success: true, data },
      { status: 200 }
    );
  } catch (error) {
    console.error('Early access form error:', error);
    return NextResponse.json(
      { error: 'Failed to submit request. Please try again.' },
      { status: 500 }
    );
  }
}
