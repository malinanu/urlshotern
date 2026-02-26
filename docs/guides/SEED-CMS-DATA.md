# Seeding Default CMS Data to Production Database

## Overview
This guide will help you populate your production database with default Page Templates, Content Blocks, Settings, and Menus, so they are readily available in your visual Page Builder.

## Prerequisites
- Access to Adminer (http://your-server:8081)
- Production database credentials (`urlshortener_db`)
- 2 minutes

## Step-by-Step Instructions

1. Open **Adminer** in your browser.
2. Login to your production database:
   - **System:** `PostgreSQL`
   - **Server**: `postgres`
   - **Username**: `urlshortener`
   - **Password**: (your DB_PASSWORD)
   - **Database**: `urlshortener_db`
3. Click on **SQL command** in the left sidebar.
4. Copy the entire contents of the newly added file: **`scripts/seed-cms-data.sql`**.
5. Paste it into the Adminer SQL text box.
6. Click **Execute**.
7. You should see a success message reading: **"Successfully seeded default CMS Page Templates, Content Blocks, Settings, and Menus."**

## What This Script Does
- Bootstraps 13 default global settings.
- Inserts 5 dynamic Database Content Blocks (e.g. Hero, Features Grid, Stats).
- Inserts 4 default Page Templates (inclusive of a Landing Page template that auto-renders the blocks).
- Adds two pre-made static pages (Home, Features).
- Adds Main Navigation and Footer link menus.

You can now refresh the `/admin` dashboard and see the Database Content Blocks and Page Templates populate automatically!
