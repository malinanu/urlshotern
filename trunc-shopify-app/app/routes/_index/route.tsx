import type { LoaderFunctionArgs } from "react-router";
import { redirect, Form, useLoaderData } from "react-router";

import { login } from "../../shopify.server";

import styles from "./styles.module.css";

export const loader = async ({ request }: LoaderFunctionArgs) => {
  const url = new URL(request.url);

  if (url.searchParams.get("shop")) {
    throw redirect(`/app?${url.searchParams.toString()}`);
  }

  return { showForm: Boolean(login) };
};

export default function App() {
  const { showForm } = useLoaderData<typeof loader>();

  return (
    <div className={styles.index}>
      <div className={styles.content}>
        <h1 className={styles.heading}>Trunc URL Shortener</h1>
        <p className={styles.text}>
          Short links for your products, collections and pages, with the clicks behind each one.
        </p>
        {showForm && (
          <Form className={styles.form} method="post" action="/auth/login">
            <label className={styles.label}>
              <span>Shop domain</span>
              <input className={styles.input} type="text" name="shop" />
              <span>e.g: my-shop-domain.myshopify.com</span>
            </label>
            <button className={styles.button} type="submit">
              Log in
            </button>
          </Form>
        )}
        <ul className={styles.list}>
          <li>
            <strong>Product links in one click</strong>. Pick a product and get a
            short link to it on your own storefront domain.
          </li>
          <li>
            <strong>See what gets clicked</strong>. Clicks by day and by country
            for every link.
          </li>
          <li>
            <strong>Your Trunc account</strong>. Connect with Sign in with Trunc;
            the links also show up in your Trunc dashboard.
          </li>
        </ul>
      </div>
    </div>
  );
}
