import type { ActionFunctionArgs } from "react-router";
import { authenticate } from "../shopify.server";
import db from "../db.server";
import { deleteShopData } from "../lib/shop.server";

// Mandatory compliance webhook, 48 hours after uninstall: delete everything for the shop.
export const action = async ({ request }: ActionFunctionArgs) => {
  const { shop } = await authenticate.webhook(request);
  await deleteShopData(db, shop);
  return new Response();
};
