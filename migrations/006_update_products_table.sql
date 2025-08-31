-- Add new columns to products table for enterprise features
ALTER TABLE products 
ADD COLUMN IF NOT EXISTS vendor_id UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS sku VARCHAR(100) UNIQUE,
ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS weight DECIMAL(10,3),
ADD COLUMN IF NOT EXISTS dimensions VARCHAR(100),
ADD COLUMN IF NOT EXISTS image_urls TEXT[],
ADD COLUMN IF NOT EXISTS tags TEXT[];

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_products_vendor_id ON products(vendor_id);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active);
CREATE INDEX IF NOT EXISTS idx_products_tags ON products USING GIN(tags);

-- Add constraint to ensure SKU is not null for active products
ALTER TABLE products ADD CONSTRAINT chk_sku_required 
    CHECK (is_active = false OR sku IS NOT NULL);