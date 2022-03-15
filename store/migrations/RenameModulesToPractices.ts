import * as admin from "firebase-admin";
import { readFileSync } from "fs";
import * as yaml from "js-yaml";

const config = yaml.safeLoad(readFileSync("../../config.yaml"));
admin.initializeApp({ credential: admin.credential.applicationDefault(), projectId: config["gcp-project"] });

const firestore = admin.firestore();

const copyModules = async () => {
    try {
        const mods = await firestore.collection("modules").get();
        console.log("Found " + mods.size + " modules");

        mods.forEach(async function(doc) {
            console.log("Migrating " + doc.id);
            const data = doc.data();
            const newData = { Practices: data["Modules"] };
            await firestore
                .collection("practices")
                .doc(doc.id)
                .set(newData);
        });
    } catch (e) {
        throw e;
    }

    console.log("modules collection copied to practices, check and delete modules via the UI");
};

const plans = async () => {
    try {
        const plans = await firestore.collection("plans").get();
        console.log("Found " + plans.size + " plans");

        for (const plan of plans.docs) {
            console.log("Migrating " + plan.id);
            const revisions = await plan.ref.collection("revisions").get();
            console.log("Found " + revisions.size + " revisions");
            for (const revision of revisions.docs) {
                const newPlan = revision.data()["Plan"];

                newPlan["Responses"]["PracticeResponses"] = newPlan["Responses"]["ModuleResponses"];
                delete newPlan["Responses"]["ModuleResponses"];

                newPlan["Responses"]["PracticesVersion"] = newPlan["Responses"]["ModulesVersion"];
                delete newPlan["Responses"]["ModulesVersion"];

                for (const pId in newPlan["Responses"]["PracticeResponses"]) {
                    const p = newPlan["Responses"]["PracticeResponses"][pId];
                    p["Practice"] = p["Module"];
                    delete p["Module"];
                }
                //console.log(JSON.stringify(newPlan, null, 2));
                await revision.ref.update({ Plan: newPlan });
            }
        }
    } catch (e) {
        throw e;
    }

    console.log("plans migrated");
};

//copyModules();
//plans();
