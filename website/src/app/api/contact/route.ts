import { Resend } from 'resend';
import { NextRequest, NextResponse } from 'next/server';
import { sanitize, isValidEmail, INPUT_LIMITS } from '../../../lib/validation';

const resend = new Resend(process.env.RESEND_API_KEY);

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { name, email, message } = body;

    // Validate required fields
    if (!name || !email || !message) {
      return NextResponse.json(
        { error: 'Missing required fields' },
        { status: 400 }
      );
    }

    // Validate input lengths
    if (name.length > INPUT_LIMITS.name) {
      return NextResponse.json(
        { error: `Name must be ${INPUT_LIMITS.name} characters or less` },
        { status: 400 }
      );
    }

    if (message.length > INPUT_LIMITS.message) {
      return NextResponse.json(
        { error: `Message must be ${INPUT_LIMITS.message} characters or less` },
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
    const safeMessage = sanitize(message);

    // Send email using Resend
    const data = await resend.emails.send({
      from: 'InfraSpec Contact Form <noreply@infraspec.sh>',
      to: 'rob@brightfame.co',
      replyTo: safeEmail,
      subject: `New Contact Form Submission from ${safeName}`,
      html: `
        <h2>New Contact Form Submission</h2>
        <p><strong>Name:</strong> ${safeName}</p>
        <p><strong>Email:</strong> ${safeEmail}</p>
        <p><strong>Message:</strong></p>
        <p>${safeMessage.replace(/\n/g, '<br>')}</p>
      `,
    });

    return NextResponse.json(
      { success: true, data },
      { status: 200 }
    );
  } catch (error) {
    console.error('Contact form error:', error);
    return NextResponse.json(
      { error: 'Failed to send message. Please try again.' },
      { status: 500 }
    );
  }
}
