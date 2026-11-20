-- CreateTable
CREATE TABLE "ShopSettings" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "shop" TEXT NOT NULL,
    "apiKey" TEXT NOT NULL DEFAULT '',
    "connectedVia" TEXT NOT NULL DEFAULT '',
    "pkceState" TEXT,
    "pkceVerifier" TEXT,
    "pkceExpiresAt" DATETIME,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "Link" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "shop" TEXT NOT NULL,
    "originalUrl" TEXT NOT NULL,
    "shortCode" TEXT NOT NULL,
    "shortUrl" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateIndex
CREATE UNIQUE INDEX "ShopSettings_shop_key" ON "ShopSettings"("shop");

-- CreateIndex
CREATE UNIQUE INDEX "ShopSettings_pkceState_key" ON "ShopSettings"("pkceState");

-- CreateIndex
CREATE INDEX "Link_shop_idx" ON "Link"("shop");

-- CreateIndex
CREATE UNIQUE INDEX "Link_shop_shortCode_key" ON "Link"("shop", "shortCode");
