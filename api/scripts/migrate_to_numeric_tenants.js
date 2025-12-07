/**
 * Migration Script: Convert String-Based Tenants to Numeric IDs with Database-Per-Tenant
 *
 * This script migrates from the old architecture:
 * - String composite tenant IDs (e.g., "Hospital-A:CardiacDept")
 * - Single shared database with tenant filtering
 *
 * To the new architecture:
 * - Numeric tenant IDs (e.g., 1001)
 * - Separate database per tenant (e.g., tenant_1001)
 * - Sections stored within tenant databases
 *
 * IMPORTANT: Backup your database before running this script!
 *
 * Usage:
 *   mongosh mongodb://localhost:27017 migrate_to_numeric_tenants.js
 */

// Configuration
const MASTER_DB_NAME = "waiting_room_master";
const OLD_DB_NAME = "waiting_room"; // Your current database name
const START_TENANT_ID = 1000;
const START_SECTION_ID = 2000;

print("=== Starting Tenant Migration ===\n");

// Connect to old database
const oldDB = db.getSiblingDB(OLD_DB_NAME);
const masterDB = db.getSiblingDB(MASTER_DB_NAME);

// Step 1: Create master database and collections
print("Step 1: Creating master database and tenant registry...");
masterDB.createCollection("tenants");
masterDB.tenants.createIndex({ "tenantId": 1 }, { unique: true });
print("✓ Master database created\n");

// Step 2: Extract unique tenants from old data
print("Step 2: Extracting unique tenants from existing data...");

// Get unique building IDs (tenants) from various collections
const uniqueBuildingIDs = new Set();

// From tenants collection
const oldTenants = oldDB.tenants.find({}).toArray();
oldTenants.forEach(tenant => {
    if (tenant.buildingId) {
        uniqueBuildingIDs.add(tenant.buildingId);
    }
});

// From system_configuration collection
const configs = oldDB.system_configuration.find({ "tenantId": { $exists: true, $ne: "" } }).toArray();
configs.forEach(config => {
    if (config.tenantId) {
        uniqueBuildingIDs.add(config.tenantId);
    }
});

// From waiting_queue collection
const queueEntries = oldDB.waiting_queue.find({ "tenantId": { $exists: true, $ne: "" } }).toArray();
queueEntries.forEach(entry => {
    if (entry.tenantId) {
        uniqueBuildingIDs.add(entry.tenantId);
    }
});

print(`✓ Found ${uniqueBuildingIDs.size} unique tenants\n`);

// Step 3: Create numeric tenant IDs and mapping
print("Step 3: Creating tenant ID mapping...");
const tenantMapping = {}; // oldBuildingId -> numericTenantId
const sectionMapping = {}; // "oldBuildingId:oldSectionId" -> numericSectionId
let nextTenantID = START_TENANT_ID;
let nextSectionID = START_SECTION_ID;

// Create mapping for each building ID
Array.from(uniqueBuildingIDs).sort().forEach((buildingId, index) => {
    const tenantID = nextTenantID++;
    tenantMapping[buildingId] = tenantID;

    print(`  ${buildingId} -> ${tenantID}`);

    // Insert into master database
    masterDB.tenants.updateOne(
        { tenantId: tenantID },
        {
            $set: {
                tenantId: tenantID,
                name: buildingId, // You may want to update these manually later
                description: `Migrated from ${buildingId}`,
                databaseName: `tenant_${tenantID}`,
                status: "active",
                updatedAt: new Date()
            },
            $setOnInsert: {
                createdAt: new Date()
            }
        },
        { upsert: true }
    );
});

print(`✓ Created ${Object.keys(tenantMapping).length} tenant mappings\n`);

// Step 4: Create section mappings
print("Step 4: Creating section mappings...");

// Get unique section IDs per building
const buildingSections = {}; // buildingId -> Set of sectionIds

// From old tenants collection
oldTenants.forEach(tenant => {
    if (tenant.buildingId && tenant.sectionId) {
        if (!buildingSections[tenant.buildingId]) {
            buildingSections[tenant.buildingId] = new Set();
        }
        buildingSections[tenant.buildingId].add(tenant.sectionId);
    }
});

// From system_configuration collection
configs.forEach(config => {
    if (config.tenantId && config.sectionId) {
        if (!buildingSections[config.tenantId]) {
            buildingSections[config.tenantId] = new Set();
        }
        buildingSections[config.tenantId].add(config.sectionId);
    }
});

// From waiting_queue collection
queueEntries.forEach(entry => {
    if (entry.tenantId && entry.sectionId) {
        if (!buildingSections[entry.tenantId]) {
            buildingSections[entry.tenantId] = new Set();
        }
        buildingSections[entry.tenantId].add(entry.sectionId);
    }
});

// Create numeric section IDs
Object.keys(buildingSections).forEach(buildingId => {
    const sections = Array.from(buildingSections[buildingId]).sort();
    sections.forEach(sectionId => {
        const compositeKey = `${buildingId}:${sectionId}`;
        const numericSectionID = nextSectionID++;
        sectionMapping[compositeKey] = numericSectionID;
        print(`  ${compositeKey} -> ${numericSectionID}`);
    });
});

print(`✓ Created ${Object.keys(sectionMapping).length} section mappings\n`);

// Step 5: Migrate data to tenant-specific databases
print("Step 5: Migrating data to tenant-specific databases...\n");

Object.keys(tenantMapping).forEach(buildingId => {
    const tenantID = tenantMapping[buildingId];
    const tenantDBName = `tenant_${tenantID}`;
    const tenantDB = db.getSiblingDB(tenantDBName);

    print(`  Migrating tenant ${buildingId} (tenantID: ${tenantID}) to ${tenantDBName}...`);

    // Create collections
    const collections = [
        "system_configuration",
        "card_readers",
        "waiting_queue",
        "sections",
        "priority_config"
    ];

    collections.forEach(collName => {
        try {
            tenantDB.createCollection(collName);
        } catch (e) {
            // Collection may already exist
        }
    });

    // Migrate system_configuration
    const tenantConfigs = oldDB.system_configuration.find({ tenantId: buildingId }).toArray();
    print(`    - Migrating ${tenantConfigs.length} configurations`);
    tenantConfigs.forEach(config => {
        const newConfig = { ...config };
        delete newConfig.tenantId; // No longer needed

        // Convert sectionId to numeric
        if (config.sectionId) {
            const compositeKey = `${buildingId}:${config.sectionId}`;
            newConfig.sectionId = sectionMapping[compositeKey] || null;
        } else {
            newConfig.sectionId = null;
        }

        tenantDB.system_configuration.insertOne(newConfig);
    });

    // Migrate card_readers
    const tenantReaders = oldDB.card_readers.find({ tenantId: buildingId }).toArray();
    print(`    - Migrating ${tenantReaders.length} card readers`);
    tenantReaders.forEach(reader => {
        const newReader = { ...reader };
        delete newReader.tenantId;

        if (reader.sectionId) {
            const compositeKey = `${buildingId}:${reader.sectionId}`;
            newReader.sectionId = sectionMapping[compositeKey] || null;
        } else {
            newReader.sectionId = null;
        }

        tenantDB.card_readers.insertOne(newReader);
    });

    // Migrate waiting_queue
    const tenantQueue = oldDB.waiting_queue.find({ tenantId: buildingId }).toArray();
    print(`    - Migrating ${tenantQueue.length} queue entries`);
    tenantQueue.forEach(entry => {
        const newEntry = { ...entry };
        delete newEntry.tenantId;

        if (entry.sectionId) {
            const compositeKey = `${buildingId}:${entry.sectionId}`;
            newEntry.sectionId = sectionMapping[compositeKey] || null;
        } else {
            newEntry.sectionId = null;
        }

        tenantDB.waiting_queue.insertOne(newEntry);
    });

    // Migrate priority_config
    if (oldDB.getCollectionNames().includes("priority_config")) {
        const tenantPriorityConfigs = oldDB.priority_config.find({ tenantId: buildingId }).toArray();
        if (tenantPriorityConfigs.length > 0) {
            print(`    - Migrating ${tenantPriorityConfigs.length} priority configurations`);
            tenantPriorityConfigs.forEach(config => {
                const newConfig = { ...config };
                delete newConfig.tenantId;

                if (config.sectionId) {
                    const compositeKey = `${buildingId}:${config.sectionId}`;
                    newConfig.sectionId = sectionMapping[compositeKey] || null;
                } else {
                    newConfig.sectionId = null;
                }

                tenantDB.priority_config.insertOne(newConfig);
            });
        }
    }

    // Create sections collection entries
    const sections = buildingSections[buildingId] || new Set();
    if (sections.size > 0) {
        print(`    - Creating ${sections.size} sections`);
        Array.from(sections).forEach(sectionId => {
            const compositeKey = `${buildingId}:${sectionId}`;
            const numericSectionID = sectionMapping[compositeKey];

            tenantDB.sections.insertOne({
                sectionId: numericSectionID,
                name: sectionId,
                description: `Migrated from ${sectionId}`,
                status: "active",
                createdAt: new Date(),
                updatedAt: new Date()
            });
        });
    }

    // Create indexes
    tenantDB.sections.createIndex({ "sectionId": 1 }, { unique: true });
    tenantDB.waiting_queue.createIndex({ "sectionId": 1, "status": 1 });
    tenantDB.waiting_queue.createIndex({ "ticketNumber": 1 });
    tenantDB.waiting_queue.createIndex({ "qrToken": 1 });
    tenantDB.card_readers.createIndex({ "id": 1, "sectionId": 1 });
    tenantDB.system_configuration.createIndex({ "sectionId": 1 });

    print(`  ✓ Completed migration for tenant ${buildingId}\n`);
});

// Step 6: Create mapping reference document
print("Step 6: Creating migration mapping reference...");
masterDB.migration_mappings.insertOne({
    migratedAt: new Date(),
    tenantMapping: tenantMapping,
    sectionMapping: sectionMapping,
    oldDatabaseName: OLD_DB_NAME,
    masterDatabaseName: MASTER_DB_NAME
});
print("✓ Migration mapping saved to master database\n");

// Step 7: Summary
print("=== Migration Summary ===");
print(`Tenants migrated: ${Object.keys(tenantMapping).length}`);
print(`Sections created: ${Object.keys(sectionMapping).length}`);
print(`Master database: ${MASTER_DB_NAME}`);
print(`\nTenant ID Range: ${START_TENANT_ID} - ${nextTenantID - 1}`);
print(`Section ID Range: ${START_SECTION_ID} - ${nextSectionID - 1}`);
print("\n=== Migration Complete ===\n");

print("IMPORTANT NEXT STEPS:");
print("1. Verify data in new tenant databases");
print("2. Update frontend to use numeric tenant IDs");
print("3. Test API endpoints with new tenant ID format");
print("4. Update any external integrations");
print("5. After verification, consider renaming/archiving old database");
print("\nOld database has NOT been deleted. You can still access it at: " + OLD_DB_NAME);
