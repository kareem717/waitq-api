# waitQ API (Incomplete)

waitQ is a hobby project I started but never fully completed due to time constraints. The app's primary purpose was to provide a simple waitlist solution for new product launches. 

**Note:** Some commits implementing key functionality might be missing, as they were lost on my local machine.

---

## Features and Tools

This project integrates several third-party services:
- **Stripe**: For billing functionality.
- **Supabase**: As the database and authentication solution.
- **AWS SES**: For sending transactional emails.

The API was built using the [Huma](https://github.com/danielgtaylor/huma) framework, which allows for easy generation of an [OpenAPI](https://github.com/OAI/OpenAPI-Specification) specification. This enabled seamless SDK client generation and integration with a [Next.js](https://github.com/vercel/next.js) frontend, while ensuring type safety.

You can find the frontend repository for this project here:  
👉 [waitQ Web Frontend](https://github.com/kareem717/waitq-web)

---

## Disclaimer

This project was a personal learning exercise, and the code quality is not production-grade. While you’re free to reuse the code, I wouldn’t recommend it as a reference for best practices.

---

## Future Plans

At the moment, I don’t have plans to complete or maintain this project. However, feel free to fork it and build upon it if you find it useful!
