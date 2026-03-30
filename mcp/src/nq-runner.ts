import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

/**
 * Executes a Gremlin query using the 'nq' binary
 */
export async function runGremlinQuery(
  query: string, 
  awsProfile?: string
): Promise<any> {
  // Escape double quotes for the shell command
  const escapedQuery = query.replace(/"/g, '\\"');
  
  let command = `nq "${escapedQuery}"`;
  if (awsProfile) {
    command += ` --aws-profile ${awsProfile}`;
  }

  try {
    const { stdout } = await execAsync(command);
    return JSON.parse(stdout);
  } catch (error: any) {
    throw new Error(`Neptune Query Failed: ${error.message}`);
  }
}

/**
 * Discovers the labels and counts in the graph
 */
export async function getGraphSchema(awsProfile?: string): Promise<string> {
  const schemaQuery = `g.V().groupCount().by(label)`;
  const result = await runGremlinQuery(schemaQuery, awsProfile);
  return JSON.stringify(result, null, 2);
}
