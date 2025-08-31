-- Create reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id),
    user_id UUID NOT NULL REFERENCES users(id),
    order_id UUID REFERENCES orders(id),
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    title VARCHAR(200) NOT NULL,
    comment TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    is_verified_purchase BOOLEAN DEFAULT false,
    helpful_count INTEGER DEFAULT 0 CHECK (helpful_count >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, user_id, order_id)
);

-- Create review_moderation table
CREATE TABLE IF NOT EXISTS review_moderation (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    review_id UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    moderator_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(20) NOT NULL CHECK (action IN ('approved', 'rejected')),
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for reviews
CREATE INDEX IF NOT EXISTS idx_reviews_product_id ON reviews(product_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_order_id ON reviews(order_id);
CREATE INDEX IF NOT EXISTS idx_reviews_rating ON reviews(rating);
CREATE INDEX IF NOT EXISTS idx_reviews_status ON reviews(status);
CREATE INDEX IF NOT EXISTS idx_reviews_is_verified ON reviews(is_verified_purchase);
CREATE INDEX IF NOT EXISTS idx_reviews_created_at ON reviews(created_at);

-- Create indexes for review_moderation
CREATE INDEX IF NOT EXISTS idx_review_moderation_review_id ON review_moderation(review_id);
CREATE INDEX IF NOT EXISTS idx_review_moderation_moderator_id ON review_moderation(moderator_id);

-- Create triggers for updating updated_at columns
CREATE TRIGGER update_reviews_updated_at BEFORE UPDATE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to update verified purchase status
CREATE OR REPLACE FUNCTION update_review_verified_purchase()
RETURNS TRIGGER AS $$
BEGIN
    -- Set is_verified_purchase to true if order_id is provided and order is delivered
    IF NEW.order_id IS NOT NULL THEN
        SELECT CASE 
            WHEN o.status = 'delivered' THEN true 
            ELSE false 
        END INTO NEW.is_verified_purchase
        FROM orders o 
        WHERE o.id = NEW.order_id AND o.user_id = NEW.user_id;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically set verified purchase status
CREATE TRIGGER set_review_verified_purchase BEFORE INSERT OR UPDATE ON reviews
    FOR EACH ROW EXECUTE FUNCTION update_review_verified_purchase();